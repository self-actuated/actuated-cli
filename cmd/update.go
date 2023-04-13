package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeUpgrade() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade an agent's kernel and root filesystem",
	}

	cmd.RunE = runUpgradeE

	cmd.Flags().StringP("owner", "o", "", "Owner")
	cmd.Flags().BoolP("staff", "s", false, "Staff flag")
	cmd.Flags().BoolP("force", "f", false, "Force upgrade")
	cmd.Flags().String("host", "", "Host to upgrade")

	return cmd
}

func runUpgradeE(cmd *cobra.Command, args []string) error {
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

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	host, err := cmd.Flags().GetString("host")
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

	res, status, err := c.UpgradeAgent(pat, owner, host, force, staff)
	if err != nil {
		return err
	}

	if status != http.StatusOK && status != http.StatusAccepted &&
		status != http.StatusNoContent && status != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d, error: %s", status, res)
	}

	fmt.Printf("Upgrade requested for %s, status: %d\n", owner, status)
	if strings.TrimSpace(res) != "" {
		fmt.Printf("Response: %s\n", res)
	}

	return nil
}
