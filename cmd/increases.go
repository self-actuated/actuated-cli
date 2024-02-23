package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeIncreases() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "increases",
		Short: "Get job increases for an organisation",
	}

	cmd.RunE = runIncreasesE

	cmd.Flags().Int("days", 30, "The number of days to look back for increases")
	cmd.Flags().BoolP("json", "j", false, "Request output in JSON format")

	return cmd
}

func runIncreasesE(cmd *cobra.Command, args []string) error {

	var owner string
	if len(args) == 1 {
		owner = strings.TrimSpace(args[0])
	}

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
	days, err := cmd.Flags().GetInt("days")
	if err != nil {
		return err
	}

	startDate := time.Now().Add(-1 * time.Duration(days) * 24 * time.Hour)

	res, status, err := c.GetBuildIncreases(pat, owner, startDate, staff, requestJson)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, message: %s", status, string(res))
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
