package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var root *cobra.Command

func init() {

	root = &cobra.Command{
		Use:   "actuated-cli",
		Short: "The official CLI for actuated",
		Long: `This CLI can be used to review and manage jobs, and the actuated
agent installed on your servers.

For GitHub:
The --owner flag or OWNER argument is a GitHub organization, i.e. for the path:
self-actuated/actuated-cli, the owner is "self-actuated" also known as an org.

For GitLab:
The NAMESPACE argument is a GitLab namespace (group), used for filtering jobs.

Run "actuated-cli auth --url URL" to authenticate with GitHub.
Run "actuated-cli auth --platform gitlab --url URL" to authenticate with GitLab.

Learn more:
https://docs.actuated.com/tasks/cli/
https://github.com/self-actuated/actuated-cli
`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.PersistentFlags().String("token-value", "", "Personal Access Token")
	root.PersistentFlags().StringP("token", "t", "$HOME/.actuated/PAT", "File to read for Personal Access Token (legacy fallback)")
	root.PersistentFlags().BoolP("staff", "s", false, "Execute the command as an actuated staff member")

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Skip URL validation for commands that don't need it
		cmdName := cmd.Name()
		if cmdName == "auth" || cmdName == "version" {
			return nil
		}

		_, err := getControllerURL()
		return err
	}

	root.AddCommand(makeAuth())
	root.AddCommand(MakeVersion())
	root.AddCommand(makeSSH())

	root.AddCommand(makeRunners())
	root.AddCommand(makeJobs())
	root.AddCommand(makeRepair())
	root.AddCommand(makeIncreases())

	root.AddCommand(makeRestart())
	root.AddCommand(makeAgentLogs())
	root.AddCommand(makeDisableAgent())
	root.AddCommand(makeUpgrade())
	root.AddCommand(makeLogs())

	root.AddCommand(makeController())
	root.AddCommand(makeMetering())
}

func Execute() error {
	return root.Execute()
}

// getPat returns the authentication token for the current controller.
// Resolution order:
// 1. --token-value flag (explicit value)
// 2. Config file entry for the current ACTUATED_URL
// 3. --token flag / legacy PAT file
func getPat(cmd *cobra.Command) (string, error) {
	// 1. Explicit --token-value flag takes highest priority
	if cmd.Flags().Changed("token-value") {
		v, err := cmd.Flags().GetString("token-value")
		if err != nil {
			return "", err
		}
		return v, nil
	}

	// 2. Try config file for the current controller URL
	cc, controllerURL, found, err := getControllerConfig()
	if err == nil && found {
		// For GitHub: use the stored token directly
		if cc.Token != "" {
			return cc.Token, nil
		}

		// For GitLab: use the cached id_token if still valid, otherwise refresh
		if cc.Platform == PlatformGitLab && cc.RefreshToken != "" {
			if cc.IDToken != "" && isIDTokenValid(cc.IDToken) {
				return cc.IDToken, nil
			}

			idToken, err := refreshOIDCToken(controllerURL, cc)
			if err != nil {
				return "", err
			}
			return idToken, nil
		}
	}

	// 3. Fall back to legacy --token flag / PAT file
	v, err := cmd.Flags().GetString("token")
	if err != nil {
		return "", err
	}

	if len(v) == 0 {
		return "", fmt.Errorf("no token found: run \"actuated-cli auth --url URL\" to authenticate, or use --token-value")
	}

	patFile := os.ExpandEnv(v)
	pat, err := readPatFile(patFile)
	if err != nil {
		return "", err
	}

	return pat, nil
}

func readPatFile(filePath string) (string, error) {
	patData, err := os.ReadFile(os.ExpandEnv(filePath))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(patData)), nil
}
