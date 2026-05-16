// Package commands wires the cobra command tree.
package commands

import (
	"github.com/spf13/cobra"

	appscmd "github.com/akshitkrnagpal/revcat/commands/apps"
	auditlogscmd "github.com/akshitkrnagpal/revcat/commands/auditlogs"
	authcmd "github.com/akshitkrnagpal/revcat/commands/auth"
	collaboratorscmd "github.com/akshitkrnagpal/revcat/commands/collaborators"
	doctorcmd "github.com/akshitkrnagpal/revcat/commands/doctor"
	entitlementscmd "github.com/akshitkrnagpal/revcat/commands/entitlements"
	initcmd "github.com/akshitkrnagpal/revcat/commands/initcmd"
	invoicescmd "github.com/akshitkrnagpal/revcat/commands/invoices"
	metricscmd "github.com/akshitkrnagpal/revcat/commands/metrics"
	offeringscmd "github.com/akshitkrnagpal/revcat/commands/offerings"
	packagescmd "github.com/akshitkrnagpal/revcat/commands/packages"
	paywallscmd "github.com/akshitkrnagpal/revcat/commands/paywalls"
	productscmd "github.com/akshitkrnagpal/revcat/commands/products"
	projectscmd "github.com/akshitkrnagpal/revcat/commands/projects"
	publishcmd "github.com/akshitkrnagpal/revcat/commands/publish"
	purchasescmd "github.com/akshitkrnagpal/revcat/commands/purchases"
	subscriberscmd "github.com/akshitkrnagpal/revcat/commands/subscribers"
	subscriptionscmd "github.com/akshitkrnagpal/revcat/commands/subscriptions"
	versioncmd "github.com/akshitkrnagpal/revcat/commands/version"
	virtualcurrenciescmd "github.com/akshitkrnagpal/revcat/commands/virtualcurrencies"
	webhookscmd "github.com/akshitkrnagpal/revcat/commands/webhooks"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

// Build metadata. All set at build time via -ldflags injection (see
// the Makefile and .goreleaser.yaml). The defaults below mark a build
// that wasn't built through either path - useful for
// `go run ./cmd/revcat` during development.
var (
	Version    = "0.0.1-dev"
	CommitHash = ""
	BuildTime  = ""
)

// Global flags. Stored at package level so subcommands can read them via
// the cmd.Root() ancestor without prop drilling.
type globalFlags struct {
	Verbose   bool
	Quiet     bool
	NoColor   bool
	Debug     bool
	Output    string
	Pretty    bool
	Profile   string
	ProjectID string
}

var Flags = &globalFlags{}

// RootCmd returns the top-level revcat command. Exposed for tools that
// introspect the cobra tree (e.g. the CLI reference doc generator under
// scripts/gen-cli-reference).
//
// Don't use this in command handlers - they get their own *cobra.Command
// via the runE callback.
func RootCmd() *cobra.Command { return rootCmd }

// rootCmd is the top-level revcat command.
var rootCmd = &cobra.Command{
	Use:   "revcat",
	Short: "The RevenueCat CLI",
	Long: `revcat is a CLI for managing RevenueCat projects from the terminal.

Imperative verbs over declarative state. JSON-first when piped, tables when
interactive. Auth uses mode-0600 config files plus optional per-repo context.`,
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
	pf.StringVar(&Flags.ProjectID, "project-id", "", "RevenueCat project id (default: REVCAT_PROJECT_ID, ./revcat.toml, or the bound profile)")

	rootCmd.AddCommand(appscmd.Cmd)
	rootCmd.AddCommand(auditlogscmd.Cmd)
	rootCmd.AddCommand(collaboratorscmd.Cmd)
	rootCmd.AddCommand(authcmd.Cmd)
	rootCmd.AddCommand(doctorcmd.Cmd)
	rootCmd.AddCommand(entitlementscmd.Cmd)
	rootCmd.AddCommand(initcmd.Cmd)
	rootCmd.AddCommand(invoicescmd.Cmd)
	rootCmd.AddCommand(metricscmd.Cmd)
	rootCmd.AddCommand(metricscmd.ChartsCmd)
	rootCmd.AddCommand(offeringscmd.Cmd)
	rootCmd.AddCommand(packagescmd.Cmd)
	rootCmd.AddCommand(paywallscmd.Cmd)
	rootCmd.AddCommand(productscmd.Cmd)
	rootCmd.AddCommand(projectscmd.Cmd)
	rootCmd.AddCommand(publishcmd.Cmd)
	rootCmd.AddCommand(purchasescmd.Cmd)
	rootCmd.AddCommand(subscriberscmd.Cmd)
	rootCmd.AddCommand(subscriptionscmd.Cmd)
	rootCmd.AddCommand(virtualcurrenciescmd.Cmd)
	rootCmd.AddCommand(webhookscmd.Cmd)
	rootCmd.AddCommand(versioncmd.Cmd(versioncmd.BuildInfo{
		Version:    Version,
		CommitHash: CommitHash,
		BuildTime:  BuildTime,
	}))
}
