package main

import (
	"github.com/self-actuated/actuated-cli/cmd"
)

func main() {

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
