package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeLogs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Fetch logs from the agent's systemd service",
	}

	cmd.RunE = runLogsE

	cmd.Flags().StringP("owner", "o", "", "List logs owned by this user")
	cmd.Flags().String("host", "", "Host or name of server as displayed in actuated")
	cmd.Flags().BoolP("staff", "s", false, "List as a staff user")
	cmd.Flags().String("id", "", "ID variable for a specific runner VM hostname")
	cmd.Flags().DurationP("age", "a", time.Minute*15, "Age of logs to fetch")

	return cmd
}

func runLogsE(cmd *cobra.Command, args []string) error {
	pat, err := cmd.Flags().GetString("pat")
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

	host, err := cmd.Flags().GetString("host")
	if err != nil {
		return err
	}

	id, err := cmd.Flags().GetString("id")

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

	res, status, err := c.GetLogs(pat, owner, host, id, age, staff)

	if err != nil {
		return err
	}

	if status != http.StatusAccepted {
		return fmt.Errorf("unexpected status code: %d, body: %s", status, res)
	}

	fmt.Println(res)

	return nil

}
