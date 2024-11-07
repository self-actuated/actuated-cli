package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeJobs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "List jobs in the build queue",
		Example: `  # Check queued and in_progress jobs for an organisation
  actuated-cli jobs ORG
  
  # Get the same result, but in JSON format
  actuated-cli jobs ORG --json
  
  # Check queued and in_progress jobs for a customer
  actuated-cli jobs --staff CUSTOMER
`,
	}

	cmd.RunE = runJobsE

	cmd.Flags().BoolP("verbose", "v", false, "Show additional columns in the output")

	cmd.Flags().BoolP("json", "j", false, "Request output in JSON format")

	cmd.Flags().BoolP("urls", "u", false, "Include the URLs to the job in the output")

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

	includeURL, err := cmd.Flags().GetBool("urls")
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
		printEvents(os.Stdout, statuses, verbose, includeURL)
	}

	return nil

}

func printEvents(w io.Writer, statuses []JobStatus, verbose, includeURL bool) {
	tabwriter := tabwriter.NewWriter(w, 0, 0, 1, ' ', tabwriter.TabIndent)
	if verbose {

		st := "JOB ID\tOWNER\tREPO\tJOB\tRUNNER\tSERVER\tSTATUS\tSTARTED\tAGE\tETA\tLABELS"
		if includeURL {
			st = st + "\tURL"
		}

		fmt.Fprintln(tabwriter, st)
	} else {
		st := "OWNER\tREPO\tJOB\tSTATUS\tAGE\tETA"
		if includeURL {
			st = st + "\tURL"
		}

		fmt.Fprintln(tabwriter, st)
	}

	var (
		totalJobs    int
		totalQueued  int
		totalRunning int
	)

	totalJobs = len(statuses)

	for _, status := range statuses {
		duration := ""

		if status.StartedAt != nil && !status.StartedAt.IsZero() {
			duration = time.Since(*status.StartedAt).Round(time.Second).String()
		}

		if status.Status == "queued" {
			totalQueued++
		} else if status.Status == "in_progress" {
			totalRunning++
		}

		eta := ""
		if status.Status != "queued" && status.AverageRuntime > time.Second*0 {
			if status.StartedAt != nil {
				runningTime := time.Since(*status.StartedAt)
				avgDuration := status.AverageRuntime
				etaV := avgDuration - runningTime
				if etaV < time.Second*0 {
					v := etaV * -1
					eta = "+" + v.Round(time.Second).String()
				} else {
					eta = etaV.Round(time.Second).String()
				}
			}
		}

		if verbose {

			line := fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
				status.JobID,
				status.Owner,
				status.Repo,
				status.JobName,
				status.RunnerName,
				status.AgentName,
				status.Status,
				duration,
				eta,
				strings.Join(status.Labels, ","))
			if includeURL {
				line = line + fmt.Sprintf("\thttps://github.com/%s/%s/runs/%d", status.Owner, status.Repo, status.JobID)
			}

			fmt.Fprintln(tabwriter, line)
		} else {
			line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s",
				status.Owner,
				status.Repo,
				status.JobName,
				status.Status,
				duration,
				eta)

			if includeURL {
				line = line + fmt.Sprintf("\thttps://github.com/%s/%s/runs/%d", status.Owner, status.Repo, status.JobID)
			}

			fmt.Fprintln(tabwriter, line)

		}
	}

	tabwriter.Flush()
	if totalJobs > 0 {

		st := "\nJOBS\tRUNNING\tQUEUED"

		fmt.Fprintln(tabwriter, st)

		fmt.Fprintf(tabwriter, "%d\t%d\t%d\n", totalJobs, totalRunning, totalQueued)
		tabwriter.Flush()

	}
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
