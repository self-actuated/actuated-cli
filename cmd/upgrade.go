package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeUpgrade() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade an agent's kernel and root filesystem",
		Example: `  # Upgrade the agent if a newer one is available
  actuated-cli upgrade --owner ORG HOST
  
  # Force an upgrade, even if on the latest version of the agent
  actuated-cli upgrade --owner ORG --force HOST
`,
	}

	cmd.RunE = runUpgradeE

	cmd.Flags().StringP("owner", "o", "", "Owner")
	cmd.Flags().BoolP("force", "f", false, "Force upgrade")
	cmd.Flags().BoolP("all", "a", false, "Upgrade all hosts instead of giving --host")

	return cmd
}

func runUpgradeE(cmd *cobra.Command, args []string) error {

	allHosts, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	if len(args) < 1 && !allHosts {
		return fmt.Errorf("specify the host as an argument")
	}

	var host string
	if !allHosts {
		host = strings.TrimSpace(args[0])
	}

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

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	if !allHosts && len(host) == 0 {
		return fmt.Errorf("--all or --host is required")
	}

	// if len(owner) == 0 {
	// 	return fmt.Errorf("owner is required")
	// }

	if len(pat) == 0 {
		return fmt.Errorf("pat is required")
	}

	c := pkg.NewClient(http.DefaultClient, os.Getenv("ACTUATED_URL"))

	var upgradeHosts []Host
	if allHosts {
		hosts, httpStatus, err := c.ListRunners(pat, owner, staff, true)
		if err != nil {
			return err
		}
		if httpStatus != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", httpStatus)
		}

		var hostsList []Host

		if err := json.Unmarshal([]byte(hosts), &hostsList); err != nil {
			return err
		}
		reachableHosts := []Host{}

		for _, h := range hostsList {
			reachableHosts = append(reachableHosts, h)
		}

		if len(reachableHosts) == 0 {
			return fmt.Errorf("no reachable hosts found")
		}
		upgradeHosts = reachableHosts

	} else {
		upgradeHosts = []Host{
			{
				Name:      host,
				Customer:  owner,
				Reachable: true,
			},
		}
	}

	for _, h := range upgradeHosts {
		st := time.Now()
		fmt.Printf("Upgrading: %s (%s)\n", h.Name, h.Customer)

		if !h.Reachable {
			fmt.Printf("Can't upgrade: %s (%s), not reachable\n", h.Name, h.Customer)
		} else if h.Status != "running" {
			fmt.Printf("Can't upgrade: %s (%s), status: %s\n", h.Name, h.Customer, h.Status)
		} else {
			res, status, err := c.UpgradeAgent(pat, h.Customer, h.Name, force, staff)
			if err != nil {
				return err
			}

			if status != http.StatusOK && status != http.StatusAccepted &&
				status != http.StatusNoContent && status != http.StatusCreated {
				return fmt.Errorf("unexpected status code: %d, error: %s", status, res)
			}

			fmt.Printf("Upgrade: %s (%s): %d (%dms)\n", h.Name, h.Customer, status, time.Since(st).Milliseconds())

			if strings.TrimSpace(res) != "" {
				fmt.Printf("Response: %s\n", res)
			}
		}
		fmt.Println("")
	}

	return nil
}

type Host struct {
	Name      string `json:"name"`
	Customer  string `json:"customer"`
	Reachable bool   `json:"reachable"`
	Status    string `json:"status"`
}
