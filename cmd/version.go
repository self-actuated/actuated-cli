// Copyright (c) arkade author(s) 2022. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"

	"github.com/morikuni/aec"
	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

func PrintASCIIArt() {
	actuatedCLILogo := aec.BlueF.Apply(actuatedCLIStr)
	fmt.Print(actuatedCLILogo)
}

func MakeVersion() *cobra.Command {
	var command = &cobra.Command{
		Use:          "version",
		Short:        "Print the version",
		Example:      `  actuated-cli version`,
		Aliases:      []string{"v"},
		SilenceUsage: false,
	}
	command.Run = func(cmd *cobra.Command, args []string) {
		PrintASCIIArt()
		if len(pkg.Version) == 0 {
			fmt.Println("Version: dev")
		} else {
			fmt.Println("Version:", pkg.Version)
		}
		fmt.Println("Git Commit:", pkg.GitCommit)
		fmt.Println()

	}
	return command
}

const actuatedCLIStr = `
░█▀█░█▀▀░▀█▀░█░█░█▀█░▀█▀░█▀▀░█▀▄
░█▀█░█░░░░█░░█░█░█▀█░░█░░█▀▀░█░█
░▀░▀░▀▀▀░░▀░░▀▀▀░▀░▀░░▀░░▀▀▀░▀▀░

Command Line Interface (CLI)

Copyright OpenFaaS Ltd 2026. All rights reserved.

`
