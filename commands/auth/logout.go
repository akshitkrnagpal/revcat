package auth

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var (
	logoutAll bool
	logoutYes bool
)

var logoutCmd = &cobra.Command{
	Use:   "logout [<profile>]",
	Short: "Remove a stored auth profile",
	Long: `Delete a stored auth profile from the keychain (or local file). Pass
--all to wipe every profile.`,
	RunE: runLogout,
}

func init() {
	logoutCmd.Flags().BoolVar(&logoutAll, "all", false, "Delete every stored profile")
	logoutCmd.Flags().BoolVarP(&logoutYes, "yes", "y", false, "Skip confirmation prompt")
}

func runLogout(cmd *cobra.Command, args []string) error {
	store, err := authstore.OpenGlobal(bypassKeychain(cmd))
	if err != nil {
		return err
	}

	if logoutAll {
		names, err := store.List()
		if err != nil {
			return err
		}
		if len(names) == 0 {
			output.Info("no profiles to remove")
			return nil
		}
		if !logoutYes {
			var ok bool
			if err := survey.AskOne(&survey.Confirm{
				Message: fmt.Sprintf("Delete all %d profiles?", len(names)),
				Default: false,
			}, &ok); err != nil {
				return err
			}
			if !ok {
				return errors.New("aborted")
			}
		}
		for _, n := range names {
			if err := store.Delete(n); err != nil {
				output.Warn("delete %q: %v", n, err)
			}
		}
		if err := authstore.ClearActive(); err != nil {
			output.Warn("clear active marker: %v", err)
		}
		output.Success("removed %d profiles", len(names))
		return nil
	}

	name := "default"
	if len(args) == 1 {
		name = args[0]
	}
	if err := store.Delete(name); err != nil {
		if errors.Is(err, authstore.ErrNoProfile) {
			return fmt.Errorf("no profile named %q", name)
		}
		return err
	}
	if active, _ := authstore.GetActive(); active == name {
		if err := authstore.ClearActive(); err != nil {
			output.Warn("clear active marker: %v", err)
		}
	}
	output.Success("removed profile %q", name)
	return nil
}
