// Package version implements `revcat version`.
package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Cmd returns the `revcat version` command. Takes the version string at
// construction so commands/root can pass through ldflags-injected values.
func Cmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print revcat version and build info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("revcat %s %s/%s %s\n", version, runtime.GOOS, runtime.GOARCH, runtime.Version())
		},
	}
}
