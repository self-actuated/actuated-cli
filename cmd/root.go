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
		Use:           "actuated-cli",
		Short:         "The actuated cli",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.PersistentFlags().String("token-value", "", "Personal Access Token")
	root.PersistentFlags().StringP("token", "t", "", "File to read for Personal Access Token")

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
}

func Execute() error {
	return root.Execute()
}

func getPat(cmd *cobra.Command) (string, error) {
	pat, err := cmd.Flags().GetString("token-value")
	if err != nil {
		return "", err
	}
	if len(pat) > 0 {
		return pat, nil
	}

	patFile, err := cmd.Flags().GetString("token")
	if err != nil {
		return "", err
	}

	if len(patFile) > 0 {
		patData, err := os.ReadFile(os.ExpandEnv(patFile))
		if err != nil {
			return "", err
		}

		pat = strings.TrimSpace(string(patData))
	}

	return pat, nil
}
