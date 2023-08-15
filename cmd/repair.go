package cmd

import (
	"encoding/json"
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
		Long: `Schedule additional VMs to repair the build queue.
Use sparingly, check the build queue to see if there is a need for 
more VMs to be launched. Then, allow ample time for the new VMs to 
pick up a job by checking the build queue again for an in_progress
status.`,
		Example: `  ## Launch VMs for queued jobs in a given organisation
  actuated repair OWNER

  ## Launch VMs for queued jobs in a given organisation for a customer
  actuated repair --staff OWNER
`,
	}

	cmd.RunE = runRepairE

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
		repairRes := RepairRes{}
		if err := json.Unmarshal([]byte(res), &repairRes); err != nil {
			return err
		}

		fmt.Printf("Requeued VMs: %d\n", repairRes.VMs)
	}

	return nil
}

type RepairRes struct {
	VMs int `json:"vms"`
}
