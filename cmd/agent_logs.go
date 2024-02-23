package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeAgentLogs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent-logs",
		Short: "Fetch logs from the agent's systemd service",
		Example: `  # Latest logs for a given host:
  actuated agent-logs --owner OWNER HOST

  # Latest logs for a given time-range
  actuated agent-logs --owner OWNER --age 1h HOST
  `,
		Aliases: []string{"service-logs"},
	}

	cmd.RunE = runAgentLogsE

	cmd.Flags().StringP("owner", "o", "", "Owner for the logs")
	cmd.Flags().DurationP("age", "a", time.Minute*15, "Age of logs to fetch")

	return cmd
}

func runAgentLogsE(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("specify the host as an argument")
	}
	host := strings.TrimSpace(args[0])

	pat, err := getPat(cmd)
	if err != nil {
		return err
	}

	staff, err := cmd.Flags().GetBool("staff")
	if err != nil {
		return err
	}

	owner, err := cmd.Flags().GetString("owner")
	if err != nil {
		return err
	}

	age, err := cmd.Flags().GetDuration("age")
	if err != nil {
		return err
	}

	if len(host) == 0 {
		return fmt.Errorf("host is required")
	}

	if len(owner) == 0 {
		return fmt.Errorf("owner is required")
	}

	if len(pat) == 0 {
		return fmt.Errorf("pat is required")
	}

	c := pkg.NewClient(http.DefaultClient, os.Getenv("ACTUATED_URL"))

	res, status, err := c.GetAgentLogs(pat, owner, host, age, staff)

	if err != nil {
		return err
	}

	if status != http.StatusAccepted {
		return fmt.Errorf("unexpected status code: %d, body: %s", status, res)
	}

	fmt.Println(res)

	return nil

}
