package cmd

import (
	"github.com/spf13/cobra"
)

func makeSSH() *cobra.Command {

	ssh := &cobra.Command{
		Use:   "ssh",
		Short: "List and connect to SSH sessions",
		Long: `List and connect to SSH sessions to explore the environment
or debug a problem with a build.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	ssh.AddCommand(makeSshList())
	ssh.AddCommand(makeSshConnect())

	return ssh
}
