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

The --owner flag or OWNER argument is a GitHub organization, i.e. for the path:
self-actuated/actuated-cli, the owner is "self-actuated" also known as an org.

Run "actuated-cli auth" to authenticate with GitHub.

Learn more:
https://docs.actuated.com/tasks/cli/
https://github.com/self-actuated/actuated-cli
`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.PersistentFlags().String("token-value", "", "Personal Access Token")
	root.PersistentFlags().StringP("token", "t", "$HOME/.actuated/PAT", "File to read for Personal Access Token")
	root.PersistentFlags().BoolP("staff", "s", false, "Execute the command as an actuated staff member")

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if _, ok := os.LookupEnv("ACTUATED_URL"); !ok {
			return fmt.Errorf(`ACTUATED_URL environment variable is not set, see the CLI tab in the dashboard for instructions`)
		}
		return nil
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

func getPat(cmd *cobra.Command) (string, error) {
	var (
		pat,
		patFile string
	)

	if cmd.Flags().Changed("token-value") {
		v, err := cmd.Flags().GetString("token-value")
		if err != nil {
			return "", err
		}
		pat = v
	} else {
		v, err := cmd.Flags().GetString("token")
		if err != nil {
			return "", err
		}

		if len(v) == 0 {
			return "", fmt.Errorf("give --token or --token-value")
		}
		patFile = os.ExpandEnv(v)
	}

	if len(patFile) > 0 {
		v, err := readPatFile(patFile)
		if err != nil {
			return "", err
		}
		pat = v
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
