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

func runGitLabJobsE(cmd *cobra.Command, args []string) error {

	var namespace string
	if len(args) == 1 {
		namespace = strings.TrimSpace(args[0])
	}

	pat, err := getPat(cmd)
	if err != nil {
		return err
	}

	requestJSON, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	if len(pat) == 0 {
		return fmt.Errorf("pat is required")
	}

	controllerURL, err := getControllerURL()
	if err != nil {
		return err
	}

	c := pkg.NewClient(http.DefaultClient, controllerURL)

	acceptJSON := true

	res, status, err := c.GitLabListJobs(pat, namespace, acceptJSON)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, message: %s", status, string(res))
	}

	if requestJSON {
		var prettyJSON bytes.Buffer
		err := json.Indent(&prettyJSON, []byte(res), "", "  ")
		if err != nil {
			return err
		}
		res = prettyJSON.String()
		fmt.Println(res)
	} else {
		var statuses []GitLabJobStatus
		if err := json.Unmarshal([]byte(res), &statuses); err != nil {
			return err
		}

		printGitLabJobs(os.Stdout, statuses)
	}

	return nil
}

func printGitLabJobs(w io.Writer, statuses []GitLabJobStatus) {
	table := tablewriter.NewWriter(w)

	table.SetHeader([]string{"JOB ID", "NAMESPACE/PROJECT", "JOB NAME", "STATUS", "RUNNER", "LABELS"})

	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)

	for _, s := range statuses {
		runner := s.RunnerName
		if runner == "" {
			runner = "-"
		}

		labels := ""
		if len(s.Labels) > 0 {
			labels = strings.Join(s.Labels, ",")
		}

		table.Append([]string{
			fmt.Sprintf("%d", s.JobID),
			fmt.Sprintf("%s/%s", s.Namespace, s.Project),
			s.JobName,
			s.Status,
			runner,
			labels,
		})
	}

	table.Render()
}

// GitLabJobStatus represents a CI job in the GitLab build queue.
type GitLabJobStatus struct {
	JobID       int64      `json:"job_id"`
	PipelineID  int64      `json:"pipeline_id"`
	NamespaceID int64      `json:"namespace_id"`
	Namespace   string     `json:"namespace"`
	ProjectID   int64      `json:"project_id"`
	Project     string     `json:"project"`
	JobName     string     `json:"job_name"`
	RunnerID    int64      `json:"runner_id,omitempty"`
	RunnerName  string     `json:"runner_name,omitempty"`
	TriggeredBy string     `json:"triggered_by,omitempty"`
	Status      string     `json:"status"`
	Labels      []string   `json:"labels,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
