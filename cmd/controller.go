package cmd

import (
	"github.com/spf13/cobra"
)

func makeController() *cobra.Command {

	controller := &cobra.Command{
		Use:           "controller",
		Short:         "Staff commands for the controller",
		Hidden:        true,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	controller.AddCommand(makeControllerLogs())

	return controller
}
