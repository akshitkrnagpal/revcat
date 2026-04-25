// Package commands wires the cobra command tree.
package commands

import (
	"github.com/spf13/cobra"

	authcmd "github.com/akshitkrnagpal/revcat/commands/auth"
	doctorcmd "github.com/akshitkrnagpal/revcat/commands/doctor"
	versioncmd "github.com/akshitkrnagpal/revcat/commands/version"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

// Version is set at build time via -ldflags.
var Version = "0.0.1-dev"

// Global flags. Stored at package level so subcommands can read them via
// the cmd.Root() ancestor without prop drilling.
type globalFlags struct {
	Verbose        bool
	Quiet          bool
	NoColor        bool
	Debug          bool
	Output         string
	Pretty         bool
	Profile        string
	BypassKeychain bool
}

var Flags = &globalFlags{}

// rootCmd is the top-level revcat command.
var rootCmd = &cobra.Command{
	Use:   "revcat",
	Short: "The RevenueCat CLI",
	Long: `revcat is a CLI for managing RevenueCat projects from the terminal.

Imperative verbs over declarative state. JSON-first when piped, tables when
interactive. Auth lives in your OS keychain.`,
	Version:           Version,
	SilenceUsage:      true,
	SilenceErrors:     true,
	DisableAutoGenTag: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		output.Configure(output.Options{
			Format:  Flags.Output,
			Pretty:  Flags.Pretty,
			NoColor: Flags.NoColor,
			Quiet:   Flags.Quiet,
			Verbose: Flags.Verbose,
		})
	},
}

// Execute runs the cobra root. main() exits 1 on non-nil error.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.BoolVarP(&Flags.Verbose, "verbose", "v", false, "Show detailed output")
	pf.BoolVarP(&Flags.Quiet, "quiet", "q", false, "Suppress non-essential output")
	pf.BoolVar(&Flags.NoColor, "no-color", false, "Disable colored output")
	pf.BoolVar(&Flags.Debug, "debug", false, "Show debug information and stack traces")
	pf.StringVar(&Flags.Output, "output", "", "Output format: table | json | csv | markdown (auto-detect when empty)")
	pf.BoolVar(&Flags.Pretty, "pretty", false, "Pretty-print JSON output")
	pf.StringVar(&Flags.Profile, "profile", "", "Auth profile name (default: REVCAT_PROFILE or 'default')")
	pf.BoolVar(&Flags.BypassKeychain, "bypass-keychain", false, "Read/write auth from ./.revcat/config.json instead of OS keychain")

	rootCmd.AddCommand(authcmd.Cmd)
	rootCmd.AddCommand(doctorcmd.Cmd)
	rootCmd.AddCommand(versioncmd.Cmd(Version))
}
