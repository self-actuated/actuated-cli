package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeMetering() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metering",
		Short: "Fetch metering from a VM",
		Long:  `Fetch the metering snapshot from a specific VM.`,

		Example: `# Get the metering snapshot from a specific VM using its hostname as the --id
actuated-cli metering --owner=OWNER --id=ID HOST

# Pipe to vmmeter for pretty-printing

actuated-cli metering --owner=OWNER --id=ID HOST | vmmeter
`,
	}

	cmd.RunE = runMeteringE

	cmd.Flags().StringP("owner", "o", "", "List logs owned by this user")
	cmd.Flags().String("id", "", "ID variable for a specific runner VM hostname")

	return cmd
}

func runMeteringE(cmd *cobra.Command, args []string) error {
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

	res, status, err := c.GetMetering(pat, owner, host, id, staff)
	if err != nil {
		return err
	}

	if status != http.StatusAccepted && status != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", status, res)
	}

	fmt.Println(res)

	return nil

}
