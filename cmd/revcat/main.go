// Package main is the revcat entrypoint.
package main

import (
	"fmt"
	"os"

	"github.com/akshitkrnagpal/revcat/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
