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
		Long: `The actuated-cli is for customers and operators to query 
the status of jobs and servers.

See the project README on GitHub for more:

https://github.com/self-actuated/actuated-cli
`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.PersistentFlags().String("token-value", "", "Personal Access Token")
	root.PersistentFlags().StringP("token", "t", "", "File to read for Personal Access Token")
	root.PersistentFlags().Bool("staff", false, "Execute the command as an actuated staff member")

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if _, ok := os.LookupEnv("ACTUATED_URL"); !ok {
			return fmt.Errorf("ACTUATED_URL environment variable is not set")
		}
		return nil
	}

	root.AddCommand(makeRunners())
	root.AddCommand(makeJobs())
	root.AddCommand(makeRepair())
	root.AddCommand(makeRestart())

	root.AddCommand(makeAgentLogs())
	root.AddCommand(makeLogs())
	root.AddCommand(makeUpgrade())

	root.AddCommand(makeSSH())

	root.AddCommand(makeAuth())
	root.AddCommand(MakeVersion())

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
	} else if cmd.Flags().Changed("token") {
		v, err := cmd.Flags().GetString("token")
		if err != nil {
			return "", err
		}
		patFile = v
	} else {
		patFile = os.ExpandEnv("$HOME/.actuated/PAT")
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
