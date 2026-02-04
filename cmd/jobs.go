package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeJobs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "List jobs in the build queue",
		Long: `List the queued and in_progress jobs for a OWNER or leave the option off
to see all your authorized organisations.

Troubleshooting:

Why may a job be "stuck" as queued?

You may have overloaded your runners so that jobs have been taken off the queue
after to save thrashing. See also "actuated-cli repair"

Why may a job be showing as in_progress for days?

GitHub's API is often inconsistent, if you open the job page, you may see it's
finished or cancelled, but still showing as running. This is a flaw in GitHub
and they tend to clean these up periodically. We can mark it as hidden on our
end if you reach out to support.
`,
		Example: `  # Check queued and in_progress jobs for your authorized orgs
  actuated-cli jobs

  # See jobs with URLs and labels
  actuated-cli jobs -v

  # See jobs for a specific organisation, if you have access to multiple:
  actuated-cli jobs ORG
  
  # Get the same result, but in JSON format
  actuated-cli jobs ORG --json
  
  # Check queued and in_progress jobs for a customer
  actuated-cli jobs --staff CUSTOMER
`,
	}

	cmd.RunE = runJobsE

	cmd.Flags().BoolP("verbose", "v", false, "Show URLs")
	cmd.Flags().BoolP("json", "j", false, "Request output in JSON format")

	return cmd
}

func runJobsE(cmd *cobra.Command, args []string) error {

	var owner string
	if len(args) == 1 {
		owner = strings.TrimSpace(args[0])
	}

	pat, err := getPat(cmd)
	if err != nil {
		return err
	}

	staff, err := cmd.Flags().GetBool("staff")
	if err != nil {
		return err
	}

	requestJson, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}

	if len(pat) == 0 {
		return fmt.Errorf("pat is required")
	}

	c := pkg.NewClient(http.DefaultClient, os.Getenv("ACTUATED_URL"))

	acceptJSON := true

	res, status, err := c.ListJobs(pat, owner, staff, acceptJSON)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, message: %s", status, string(res))
	}

	if requestJson {

		var prettyJSON bytes.Buffer
		err := json.Indent(&prettyJSON, []byte(res), "", "  ")
		if err != nil {
			return err
		}
		res = prettyJSON.String()
		fmt.Println(res)
	} else {

		var statuses []JobStatus

		if err := json.Unmarshal([]byte(res), &statuses); err != nil {
			return err
		}

		printEvents(os.Stdout, statuses, verbose)
	}

	return nil

}

// progressBar generates a progress bar using soft block characters
// width is the total width of the bar, progress is 0.0 to 1.0+
// If progress > 1.0 (overflow), shows a + at the end
func progressBar(progress float64, width int) string {
	if width <= 0 {
		width = 10
	}

	if progress < 0 {
		progress = 0
	}

	// Check for overflow (job running longer than expected)
	if progress > 1.0 {
		// Full bar with + to indicate overflow
		bar := strings.Repeat("█", width-1) + "+"
		return "[" + bar + "]"
	}

	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}

	// Use soft block characters: █ for filled, ░ for empty
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return "[" + bar + "]"
}

func printEvents(w io.Writer, statuses []JobStatus, verbose bool) {
	table := tablewriter.NewWriter(w)

	// Set up headers - ETA column shows status implicitly (Queued or progress bar)
	if verbose {
		table.SetHeader([]string{"OWNER/REPO", "JOB/WORKFLOW", "RUNNER/SERVER", "ETA", "LABELS", "URL"})
	} else {
		table.SetHeader([]string{"OWNER/REPO", "JOB/WORKFLOW", "RUNNER/SERVER", "ETA", "LABELS"})
	}

	// Configure table style
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	table.SetRowLine(true)

	for _, status := range statuses {
		// Line 1 values
		owner := status.Owner + "/"
		job := status.JobName
		runner := status.RunnerName
		etaLine1 := ""
		etaLine2 := ""

		// Line 2 values
		repo := status.Repo
		workflow := status.WorkflowName
		server := status.AgentName

		if status.Status == "queued" {
			// Queued jobs show "Queued" with empty progress bar
			etaLine1 = "Queued"
			etaLine2 = progressBar(0, 10)
		} else if status.AverageRuntime > time.Second*0 && status.StartedAt != nil {
			// Running jobs with ETA data
			runningTime := time.Since(*status.StartedAt)
			avgDuration := status.AverageRuntime
			etaV := avgDuration - runningTime

			var progress float64
			if avgDuration > 0 {
				progress = float64(runningTime) / float64(avgDuration)
			}
			etaLine2 = progressBar(progress, 10)

			etaSec := int(etaV.Seconds())
			if etaSec < 0 {
				etaLine1 = fmt.Sprintf("+%ds", -etaSec)
			} else {
				etaLine1 = fmt.Sprintf("%ds", etaSec)
			}
		} else {
			// Running but no ETA data available
			etaLine1 = "Running"
			etaLine2 = ""
		}

		url := fmt.Sprintf("https://github.com/%s%s/runs/%d", owner, repo, status.JobID)
		labels := ""
		if len(status.Labels) > 0 {
			labels = strings.Join(status.Labels, ",")
		}

		// Use newlines to create two-line cells
		if verbose {
			table.Append([]string{
				owner + "\n" + repo,
				job + "\n" + workflow,
				runner + "\n" + server,
				etaLine1 + "\n" + etaLine2,
				labels,
				url,
			})
		} else {
			table.Append([]string{
				owner + "\n" + repo,
				job + "\n" + workflow,
				runner + "\n" + server,
				etaLine1 + "\n" + etaLine2,
				labels,
			})
		}
	}

	table.Render()
}

type JobStatus struct {
	JobID        int64  `json:"job_id"`
	Owner        string `json:"owner"`
	Repo         string `json:"repo"`
	WorkflowName string `json:"workflow_name"`
	JobName      string `json:"job_name"`
	Actor        string `json:"actor,omitempty"`

	RunnerName string   `json:"runner_name,omitempty"`
	Status     string   `json:"status"`
	Conclusion string   `json:"conclusion,omitempty"`
	Labels     []string `json:"labels,omitempty"`

	UpdatedAt   *time.Time `json:"updated_at"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`

	AgentName string `json:"agent_name,omitempty"`

	AverageRuntime time.Duration `json:"averageRuntime,omitempty"`

	QueuedAt *time.Time `json:"queuedAt,omitempty"`
}
