package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func runGitLabRunnersE(cmd *cobra.Command, args []string) error {

	images, err := cmd.Flags().GetBool("images")
	if err != nil {
		return err
	}

	pat, err := getPat(cmd)
	if err != nil {
		return err
	}

	requestJSON, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	if len(pat) == 0 {
		return fmt.Errorf("pat is required")
	}

	controllerURL, err := getControllerURL()
	if err != nil {
		return err
	}

	c := pkg.NewClient(http.DefaultClient, controllerURL)

	res, status, err := c.GitLabListRunners(pat, images, requestJSON)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, message: %s", status, res)
	}

	if requestJSON {
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
