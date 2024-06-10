package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeDisable() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable the actuated service remotely.",
		Example: `  # Disable the actuated systemd service from restarting

  actuated-cli disable --owner ORG HOST
`,
	}

	cmd.RunE = runDisableE

	cmd.Flags().StringP("owner", "o", "", "Owner")

	return cmd
}

func runDisableE(cmd *cobra.Command, args []string) error {
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

	if len(owner) == 0 {
		return fmt.Errorf("owner is required")
	}

	if len(pat) == 0 {
		return fmt.Errorf("pat is required")
	}

	c := pkg.NewClient(http.DefaultClient, os.Getenv("ACTUATED_URL"))

	res, status, err := c.DisableAgent(pat, owner, host, staff)
	if err != nil {
		return err
	}

	if status != http.StatusOK && status != http.StatusAccepted &&
		status != http.StatusNoContent && status != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d, error: %s", status, res)
	}

	fmt.Printf("Disable requested for %s, status: %d\n", owner, status)
	if strings.TrimSpace(res) != "" {
		fmt.Printf("Response: %s\n", res)
	}

	return nil
}
