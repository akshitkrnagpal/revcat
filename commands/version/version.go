// Package version implements `revcat version`.
package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// BuildInfo carries the ldflags-injected metadata. Empty fields are
// suppressed from the output so `go run ./cmd/revcat version` during
// development stays readable.
type BuildInfo struct {
	Version    string
	CommitHash string
	BuildTime  string
}

// Cmd returns the `revcat version` command.
func Cmd(info BuildInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print revcat version and build info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("revcat %s %s/%s %s\n",
				info.Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
			if info.CommitHash != "" {
				fmt.Printf("  commit: %s\n", info.CommitHash)
			}
			if info.BuildTime != "" {
				fmt.Printf("  built:  %s\n", info.BuildTime)
			}
		},
	}
}
