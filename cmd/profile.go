package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeProfile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile OWNER",
		Short: "View profiling snapshots for completed jobs",
		Args:  cobra.ExactArgs(1),
		Example: `  # Show the 20 most recent profiling snapshots
  actuated-cli profile openfaasltd

  # Show more snapshots or return the raw JSON
  actuated-cli profile openfaasltd --limit 50
  actuated-cli profile openfaasltd --json

  # Inspect the full snapshot for a GitHub Actions job
  actuated-cli profile openfaasltd --id 96357108346`,
		RunE: runProfileE,
	}

	cmd.Flags().IntP("limit", "l", 20, "Number of recent snapshots to show")
	cmd.Flags().String("id", "", "GitHub Actions job ID for a detailed snapshot")
	cmd.Flags().BoolP("json", "j", false, "Request output in JSON format")

	return cmd
}

func runProfileE(cmd *cobra.Command, args []string) error {
	owner := strings.TrimSpace(args[0])
	if owner == "" {
		return fmt.Errorf("owner is required")
	}

	pat, err := getPat(cmd)
	if err != nil {
		return err
	}
	if pat == "" {
		return fmt.Errorf("pat is required")
	}

	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return err
	}
	if limit < 1 {
		return fmt.Errorf("limit must be greater than zero")
	}

	jobID, err := cmd.Flags().GetString("id")
	if err != nil {
		return err
	}
	if jobID != "" {
		if _, err := strconv.ParseInt(jobID, 10, 64); err != nil {
			return fmt.Errorf("id must be a numeric GitHub Actions job ID")
		}
	}

	staff, err := cmd.Flags().GetBool("staff")
	if err != nil {
		return err
	}
	requestJSON, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	client := pkg.NewClient(http.DefaultClient, os.Getenv("ACTUATED_URL"))
	var body string
	var status int
	if jobID == "" {
		body, status, err = client.ListProfiles(pat, owner, limit, staff)
	} else {
		body, status, err = client.GetProfile(pat, owner, jobID, staff)
	}
	if err != nil {
		return err
	}
	if status == http.StatusUnauthorized {
		return fmt.Errorf("GitHub token rejected, run \"actuated-cli auth\" and try again")
	}
	if status != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, message: %s", status, body)
	}

	if requestJSON {
		return printProfileJSON(os.Stdout, []byte(body))
	}

	if jobID != "" {
		var snapshot ProfileSnapshot
		if err := json.Unmarshal([]byte(body), &snapshot); err != nil {
			return err
		}
		printProfileDetails(os.Stdout, snapshot)
		return nil
	}

	var snapshots []ProfileSummary
	if err := json.Unmarshal([]byte(body), &snapshots); err != nil {
		return err
	}
	printProfiles(os.Stdout, snapshots)

	return nil
}

func printProfileJSON(w io.Writer, body []byte) error {
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, body, "", "  "); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, formatted.String())
	return err
}

func printProfiles(w io.Writer, snapshots []ProfileSummary) {
	table := newProfileTable(w)
	table.SetHeader([]string{"AGE", "REPO", "JOB/WORKFLOW", "AGENT/VM", "RUNTIME", "VCPU", "LOAD 1/5", "RAM", "DISK"})

	for _, snapshot := range snapshots {
		table.Append([]string{
			profileAge(snapshot.CompletedAt),
			snapshot.Repo,
			snapshot.Job + "\n" + snapshot.Workflow,
			profileAgent(snapshot.AgentName) + "\n" + valueOrNA(snapshot.Hostname),
			profileRuntime(snapshot.StartedAt, snapshot.CompletedAt),
			profileCPU(snapshot.TotalCPU, snapshot.ShareFactor),
			profileLoads(snapshot.MaxLoadAvg1, snapshot.MaxLoadAvg5),
			profileUsage(snapshot.TotalMemoryGB, snapshot.MinAvailableMemoryGB),
			profileUsage(snapshot.DiskSpaceTotalGB, snapshot.DiskSpaceFreeGB),
		})
	}

	table.Render()
}

func printProfileDetails(w io.Writer, snapshot ProfileSnapshot) {
	table := newProfileTable(w)
	table.SetHeader([]string{"FIELD", "VALUE"})
	rows := [][]string{
		{"Job ID", snapshot.JobID},
		{"Repository", snapshot.Owner + "/" + snapshot.Repo},
		{"Job", snapshot.Job},
		{"Workflow", snapshot.Workflow},
		{"Agent", profileAgent(snapshot.AgentName)},
		{"VM", valueOrNA(snapshot.Hostname)},
		{"Runtime", profileRuntime(snapshot.StartedAt, snapshot.CompletedAt)},
		{"vCPU", profileCPU(snapshot.TotalCPU, snapshot.ShareFactor)},
		{"Load average 1/5/15m", joinProfileValues(snapshot.MaxLoadAvg1, snapshot.MaxLoadAvg5, snapshot.MaxLoadAvg15)},
		{"RAM total", formatProfileBytes(snapshot.TotalMemoryBytes)},
		{"RAM minimum available", formatProfileBytes(snapshot.MinAvailableMemoryBytes)},
		{"Disk total/free/used", joinProfileBytes(snapshot.DiskSpaceTotalBytes, snapshot.DiskSpaceFreeBytes, snapshot.DiskSpaceUsedBytes)},
		{"Disk read/write", joinProfileBytes(snapshot.DiskReadTotalBytes, snapshot.DiskWriteTotalBytes)},
		{"Network RX/TX", joinProfileBytes(snapshot.NetworkReadTotalBytes, snapshot.NetworkWriteTotalBytes)},
	}
	for _, row := range rows {
		table.Append(row)
	}
	table.Render()
}

func newProfileTable(w io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(w)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	return table
}

func profileAge(completed *time.Time) string {
	if completed == nil {
		return "n/a"
	}
	age := time.Since(*completed)
	if age < 0 {
		age = 0
	}
	return age.Round(time.Second).String()
}

func profileAgent(agentName string) string {
	if agentName == "" {
		return "n/a"
	}
	return agentName
}

func profileRuntime(startedAt, completedAt *time.Time) string {
	if startedAt == nil || completedAt == nil {
		return "n/a"
	}
	return completedAt.Sub(*startedAt).Round(time.Second).String()
}

func profileCPU(totalCPU *int, shareFactor *float64) string {
	if totalCPU == nil {
		return "n/a"
	}
	cpu := float64(*totalCPU)
	if shareFactor != nil && *shareFactor > 0 {
		cpu *= *shareFactor
	}
	return strconv.FormatFloat(cpu, 'f', -1, 64)
}

func profileLoads(load1, load5 *float64) string {
	return joinProfileValues(load1, load5)
}

func profileUsage(total, available *float64) string {
	if total == nil || available == nil || *total <= 0 ||
		math.IsNaN(*total) || math.IsNaN(*available) ||
		math.IsInf(*total, 0) || math.IsInf(*available, 0) {
		return "n/a"
	}
	used := *total - *available
	percentage := used / *total * 100
	return fmt.Sprintf("%.2f/%.2fGB (%.0f%%)", used, *total, percentage)
}

func valueOrNA(value string) string {
	if value == "" {
		return "n/a"
	}
	return value
}

func joinProfileValues(values ...*float64) string {
	parts := make([]string, len(values))
	for i, value := range values {
		if value == nil {
			parts[i] = "n/a"
		} else {
			parts[i] = strconv.FormatFloat(*value, 'f', -1, 64)
		}
	}
	return strings.Join(parts, "/")
}

func joinProfileBytes(values ...*float64) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = formatProfileBytes(value)
	}
	return strings.Join(parts, "/")
}

func formatProfileBytes(value *float64) string {
	if value == nil {
		return "n/a"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := *value
	unit := 0
	for math.Abs(size) >= 1000 && unit < len(units)-1 {
		size /= 1000
		unit++
	}
	return fmt.Sprintf("%.2f%s", size, units[unit])
}

// ProfileSummary matches the controller's snapshots list response. Capacity
// fields are converted to decimal GB by the controller query.
type ProfileSummary struct {
	JobID       string     `json:"job_id"`
	Owner       string     `json:"owner"`
	Repo        string     `json:"repo"`
	Job         string     `json:"job"`
	Workflow    string     `json:"workflow"`
	AgentName   string     `json:"agent_name,omitempty"`
	Hostname    string     `json:"hostname,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	TotalCPU             *int     `json:"total_cpu,omitempty"`
	TotalMemoryGB        *float64 `json:"total_memory,omitempty"`
	MinAvailableMemoryGB *float64 `json:"min_memory_available_gb,omitempty"`
	ShareFactor          *float64 `json:"share_factor,omitempty"`
	MaxLoadAvg1          *float64 `json:"max_load_avg1,omitempty"`
	MaxLoadAvg5          *float64 `json:"max_load_avg5,omitempty"`
	DiskSpaceTotalGB     *float64 `json:"disk_space_total,omitempty"`
	DiskSpaceFreeGB      *float64 `json:"disk_space_free,omitempty"`
}

// ProfileSnapshot matches the controller's detailed snapshot response. Byte
// counters are returned directly from the stored vmmeter snapshot.
type ProfileSnapshot struct {
	JobID       string     `json:"job_id"`
	Owner       string     `json:"owner"`
	Repo        string     `json:"repo"`
	Job         string     `json:"job"`
	Workflow    string     `json:"workflow"`
	AgentName   string     `json:"agent_name,omitempty"`
	Hostname    string     `json:"hostname,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	TotalCPU                *int     `json:"total_cpu,omitempty"`
	TotalMemoryBytes        *float64 `json:"total_memory,omitempty"`
	MinAvailableMemoryBytes *float64 `json:"min_memory_available_gb,omitempty"`
	ShareFactor             *float64 `json:"share_factor,omitempty"`
	MaxLoadAvg1             *float64 `json:"max_load_avg1,omitempty"`
	MaxLoadAvg5             *float64 `json:"max_load_avg5,omitempty"`
	MaxLoadAvg15            *float64 `json:"max_load_avg15,omitempty"`
	DiskSpaceTotalBytes     *float64 `json:"disk_space_total,omitempty"`
	DiskSpaceFreeBytes      *float64 `json:"disk_space_free,omitempty"`
	DiskSpaceUsedBytes      *float64 `json:"disk_space_used,omitempty"`
	DiskReadTotalBytes      *float64 `json:"disk_read_total,omitempty"`
	DiskWriteTotalBytes     *float64 `json:"disk_write_total,omitempty"`
	NetworkReadTotalBytes   *float64 `json:"network_read_total,omitempty"`
	NetworkWriteTotalBytes  *float64 `json:"network_write_total,omitempty"`
}
