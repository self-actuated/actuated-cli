package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func makeRunners() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runners",
		Short: "List runners for an organisation",
		Example: ` # List runners for a given organisation
  actuated-cli runners OWNER
  
  # List runners for all customers
  actuated-cli runners --staff OWNER

  # List runners in JSON format
  actuated-cli runners --json OWNER
`,
	}

	cmd.RunE = runRunnersE

	cmd.Flags().Bool("images", false, "Show the image being used for the rootfs and Kernel")
	cmd.Flags().BoolP("json", "j", false, "Request output in JSON format")

	return cmd
}

func runRunnersE(cmd *cobra.Command, args []string) error {

	platform := getPlatform()
	if platform == PlatformGitLab {
		return runGitLabRunnersE(cmd, args)
	}

	images, err := cmd.Flags().GetBool("images")
	if err != nil {
		return err
	}

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

	controllerURL, err := getControllerURL()
	if err != nil {
		return err
	}

	c := pkg.NewClient(http.DefaultClient, controllerURL)

	res, status, err := c.ListRunners(pat, owner, staff, images, requestJson)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, message: %s", status, res)
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
