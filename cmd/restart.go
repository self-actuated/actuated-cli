package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeRestart() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Forcibly restart the agent by killing it or reboot the machine.",
		Example: `  # Request the agent to restart
  # This will drain any running jobs - do a forced upgrade if you want to 
  # restart gracefully.
  actuated-cli restart --owner ORG HOST

  # Reboot the machine, if the agent is not responding.
  # This will not drain any running jobs.
  actuated-cli restart --owner ORG --reboot HOST
`,
	}

	cmd.RunE = runRestartE

	cmd.Flags().StringP("owner", "o", "", "Owner")
	cmd.Flags().BoolP("reboot", "r", false, "Reboot the machine instead of just restarting the service")

	return cmd
}

func runRestartE(cmd *cobra.Command, args []string) error {
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

	reboot, err := cmd.Flags().GetBool("reboot")
	if err != nil {
		return err
	}

	if len(owner) == 0 {
		return fmt.Errorf("owner is required")
	}

	if len(pat) == 0 {
		return fmt.Errorf("pat is required")
	}

	c := pkg.NewClient(http.DefaultClient, os.Getenv("ACTUATED_URL"))

	res, status, err := c.RestartAgent(pat, owner, host, reboot, staff)
	if err != nil {
		return err
	}

	if status != http.StatusOK && status != http.StatusAccepted &&
		status != http.StatusNoContent && status != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d, error: %s", status, res)
	}

	fmt.Printf("Restart requested for %s, status: %d\n", owner, status)
	if strings.TrimSpace(res) != "" {
		fmt.Printf("Response: %s\n", res)
	}

	return nil
}
