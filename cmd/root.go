package cmd

import (
	"fmt"
	"os"

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

	// add global flag for PAT
	root.PersistentFlags().StringP("pat", "p", "", "Personal Access Token")
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if _, ok := os.LookupEnv("ACTUATED_URL"); !ok {
			return fmt.Errorf("ACTUATED_URL environment variable is not set")
		}
		return nil
	}

	root.AddCommand(makeRunners())
	root.AddCommand(makeJobs())
	root.AddCommand(makeRepair())
	root.AddCommand(makeAgentLogs())
	root.AddCommand(makeLogs())
	root.AddCommand(makeUpgrade())
}

func Execute() error {
	return root.Execute()
}
