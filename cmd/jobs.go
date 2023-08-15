package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeJobs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "List jobs in the build queue",
	}

	cmd.RunE = runJobsE

	cmd.Flags().StringP("owner", "o", "", "List jobs owned by this user")
	cmd.Flags().Bool("json", false, "Request output in JSON format")

	return cmd
}

func runJobsE(cmd *cobra.Command, args []string) error {
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

	requestJson, err := cmd.Flags().GetBool("json")
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
