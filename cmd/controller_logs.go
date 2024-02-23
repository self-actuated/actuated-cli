package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeControllerLogs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Fetch logs from the controller's systemd service",
		Example: `  # Latest logs for a given host:
  actuated controller logs --age 15m
  `,
	}

	cmd.RunE = runControllerLogsE

	cmd.Flags().DurationP("age", "a", time.Minute*15, "Age of logs to fetch")
	cmd.Flags().StringP("output", "o", "cat", "Output format, use \"cat\" for brevity")

	return cmd
}

func runControllerLogsE(cmd *cobra.Command, args []string) error {

	pat, err := getPat(cmd)
	if err != nil {
		return err
	}

	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
	age, err := cmd.Flags().GetDuration("age")
	if err != nil {
		return err
	}

	if len(pat) == 0 {
		return fmt.Errorf("pat is required")
	}

	c := pkg.NewClient(http.DefaultClient, os.Getenv("ACTUATED_URL"))

	res, status, err := c.GetControllerLogs(pat, outputFormat, age)

	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", status, res)
	}

	fmt.Println(res)

	return nil

}
