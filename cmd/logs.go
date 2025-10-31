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

func makeLogs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Fetch logs from VMs",
		Long: `Fetch logs from a specific VM or a all VMs over a 
range of time.`,

		Example: `  # Default to get logs from all VMs from past 15 mins
  actuated-cli logs --owner OWNER HOST		
		
  # Logs from all VMs over the past 1 hour
  actuated-cli logs --owner OWNER --age 1h HOST

  # Get the logs from a specific VM using its hostname as the --id
  actuated-cli logs --owner OWNER --id ID HOST
`,
	}

	cmd.RunE = runLogsE
	cmd.PreRunE = preRunLogsE
	cmd.Aliases = []string{"vm-logs", "l"}

	cmd.Flags().StringP("owner", "o", "", "List logs owned by this user")
	cmd.Flags().String("id", "", "ID variable for a specific runner VM hostname")
	cmd.Flags().DurationP("age", "a", time.Minute*15, "Age of logs to fetch specified as a Go duration")

	return cmd
}

func preRunLogsE(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("specify the host as an argument")
	}

	ageChanged := cmd.Flags().Changed("age")
	idChanged := cmd.Flags().Changed("id")

	if ageChanged && idChanged {
		return fmt.Errorf("--age can't be used with --id")
	}

	return nil

}

func runLogsE(cmd *cobra.Command, args []string) error {
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

	id, err := cmd.Flags().GetString("id")
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
