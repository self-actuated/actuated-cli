package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeRepair() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repair",
		Short: "Schedule additional VMs to repair the build queue",
	}

	cmd.RunE = runRepairE

	cmd.Flags().StringP("owner", "o", "", "List repair owned by this user")
	cmd.Flags().BoolP("staff", "s", false, "List staff repair")

	return cmd
}

func runRepairE(cmd *cobra.Command, args []string) error {
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

	res, status, err := c.Repair(pat, owner, staff)
	if err != nil {
		return err
	}

	if status != http.StatusOK && status != http.StatusAccepted &&
		status != http.StatusNoContent && status != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", status)
	}

	fmt.Printf("Repair requested for %s, status: %d\n", owner, status)
	if strings.TrimSpace(res) != "" {
		fmt.Printf("Response: %s\n", res)
	}

	return nil
}
