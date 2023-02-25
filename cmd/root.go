package cmd

import (
	"github.com/spf13/cobra"
)

var root *cobra.Command

func init() {

	root = &cobra.Command{
		Use:   "actuated-cli",
		Short: "The actuated cli",
	}

	// add global flag for PAT
	root.PersistentFlags().StringP("pat", "p", "", "Personal Access Token")

	root.AddCommand(makeRunners())
	root.AddCommand(makeJobs())
	root.AddCommand(makeRepair())

}

func Execute() error {
	return root.Execute()
}
