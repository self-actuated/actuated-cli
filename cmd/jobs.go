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

func printEvents(w io.Writer, statuses []JobStatus, verbose bool) {
	tabwriter := tabwriter.NewWriter(w, 0, 0, 1, ' ', tabwriter.TabIndent)
	if verbose {
		fmt.Fprintf(tabwriter, "JOB ID\tOWNER\tREPO\tJOB\tRUNNER\tSERVER\tSTATUS\tSTARTED\tAGE\tETA\tLABELS\tURL\n")
	} else {
		fmt.Fprintf(tabwriter, "OWNER\tREPO\tJOB\tSTATUS\tAGE\tETA\tURL\n")
	}

	for _, status := range statuses {
		duration := ""

		if status.StartedAt != nil && !status.StartedAt.IsZero() {
			duration = time.Since(*status.StartedAt).Round(time.Second).String()
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
			fmt.Fprintf(tabwriter, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				status.JobID,
				status.Owner,
				status.Repo,
				status.JobName,
				status.RunnerName,
				status.AgentName,
				status.Status,
				status.StartedAt.Format(time.RFC3339),
				duration,
				eta,
				strings.Join(status.Labels, ","),
				fmt.Sprintf("https://github.com/%s/%s/runs/%d", status.Owner, status.Repo, status.JobID),
			)
		} else {
			fmt.Fprintf(tabwriter, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				status.Owner,
				status.Repo,
				status.JobName,
				status.Status,
				duration,
				eta,
				fmt.Sprintf("https://github.com/%s/%s/runs/%d", status.Owner, status.Repo, status.JobID))

		}

	}
	tabwriter.Flush()
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
