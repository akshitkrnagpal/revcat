package auth

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var useCmd = &cobra.Command{
	Use:   "use <profile>",
	Short: "Set the default auth profile",
	Long: `Set the active auth profile by writing the name to ~/.revcat/active.
Equivalent to passing --profile <name> on every command, or setting
REVCAT_PROFILE in your shell.`,
	Args: cobra.ExactArgs(1),
	RunE: runUse,
}

func runUse(cmd *cobra.Command, args []string) error {
	store, err := authstore.OpenGlobal()
	if err != nil {
		return err
	}
	if _, err := store.Get(args[0]); err != nil {
		if errors.Is(err, authstore.ErrNoProfile) {
			return fmt.Errorf("no profile named %q. run `revcat auth list` to see available profiles", args[0])
		}
		return err
	}
	if err := authstore.SetActive(args[0]); err != nil {
		return err
	}
	output.Success("active profile set to %q", args[0])
	output.Info("(persisted to ~/.revcat/active; override with --profile or REVCAT_PROFILE)")
	return nil
}
