package cmd

import (
	"github.com/spf13/cobra"
)

func makeSSH() *cobra.Command {

	ssh := &cobra.Command{
		Use:           "ssh",
		Short:         "Manage SSH sessions",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	ssh.AddCommand(makeSshList())
	ssh.AddCommand(makeSshConnect())

	return ssh
}
