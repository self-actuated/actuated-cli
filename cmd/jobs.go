package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeJobs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "List jobs in the build queue",
		Example: `  # Check queued and in_progress jobs for an organisation
  actuated-cli jobs ORG
  
  # Get the same result, but in JSON format
  actuated-cli jobs ORG --json
  
  # Check queued and in_progress jobs for a customer
  actuated-cli jobs --staff CUSTOMER
`,
	}

	cmd.RunE = runJobsE

	cmd.Flags().Bool("json", false, "Request output in JSON format")

	return cmd
}

func runJobsE(cmd *cobra.Command, args []string) error {

	if len(args) < 1 {
		return fmt.Errorf("give an owner as an argument")
	}
	owner := strings.TrimSpace(args[0])

	pat, err := getPat(cmd)
	if err != nil {
		return err
	}

	staff, err := cmd.Flags().GetBool("staff")
	if err != nil {
		return err
	}

	requestJson, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	if len(pat) == 0 {
		return fmt.Errorf("pat is required")
	}

	c := pkg.NewClient(http.DefaultClient, os.Getenv("ACTUATED_URL"))

	res, status, err := c.ListJobs(pat, owner, staff, requestJson)

	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", status)
	}

	if requestJson {

		var prettyJSON bytes.Buffer
		err := json.Indent(&prettyJSON, []byte(res), "", "  ")
		if err != nil {
			return err
		}
		res = prettyJSON.String()
	}
	fmt.Println(res)

	return nil

}
