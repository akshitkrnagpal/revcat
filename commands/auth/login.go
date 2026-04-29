package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var (
	loginName        string
	loginSecretKey   string
	loginSecretStdin bool
	loginProjectID   string
	loginNoVerify    bool
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save a RevenueCat secret key as a named profile",
	Long: `Save a RevenueCat v2 secret key (sk_...) as a named profile.

Without flags this is interactive. With flags it's scriptable. The key is
written to your OS keychain unless --bypass-keychain is set, in which case
it's written to ./.revcat/config.json (a .gitignore is created on first use).

Examples:
  revcat auth login                                                   # interactive
  echo $RC_KEY | revcat auth login --name prod --secret-key-stdin     # scripted, no shell history
  revcat auth login --name prod --secret-key sk_xxx                   # scripted (key visible in history)
  echo $RC_KEY | revcat auth login --name ci --secret-key-stdin \
    --project-id proj_xxx --no-verify                                 # CI`,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().StringVarP(&loginName, "name", "n", "", "Profile name (default: 'default')")
	loginCmd.Flags().StringVarP(&loginSecretKey, "secret-key", "k", "", "RevenueCat v2 secret key (sk_...). Warning: visible in shell history; prefer --secret-key-stdin in production")
	loginCmd.Flags().BoolVar(&loginSecretStdin, "secret-key-stdin", false, "Read the secret key from stdin (recommended for scripts/CI; avoids leaking the key into shell history)")
	loginCmd.Flags().StringVar(&loginProjectID, "project-id", "", "Project id to bind (auto-detected from /v2/projects if omitted)")
	loginCmd.Flags().BoolVar(&loginNoVerify, "no-verify", false, "Skip the API check that the key is valid")
}

func runLogin(cmd *cobra.Command, args []string) error {
	if loginName == "" {
		loginName = "default"
	}
	if loginSecretStdin && loginSecretKey != "" {
		return fmt.Errorf("--secret-key and --secret-key-stdin are mutually exclusive; pass only one")
	}
	if loginSecretStdin {
		raw, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("read secret key from stdin: %w", err)
		}
		loginSecretKey = strings.TrimSpace(string(raw))
		if loginSecretKey == "" {
			return fmt.Errorf("--secret-key-stdin was set but stdin was empty")
		}
	}
	if loginSecretKey == "" {
		if err := survey.AskOne(&survey.Password{Message: "RevenueCat secret key (sk_...)"}, &loginSecretKey); err != nil {
			return err
		}
	}
	loginSecretKey = strings.TrimSpace(loginSecretKey)
	if !strings.HasPrefix(loginSecretKey, "sk_") {
		output.Warn("secret keys usually start with sk_; got %q. proceeding anyway.", redactKey(loginSecretKey))
	}

	store, err := authstore.Open(bypassKeychain(cmd))
	if err != nil {
		return err
	}

	projectID := loginProjectID
	if !loginNoVerify {
		client := api.New(api.Options{SecretKey: loginSecretKey, ProjectID: projectID, Version: cmd.Root().Version})
		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()
		projects, err := client.ListProjects(ctx)
		if err != nil {
			var apiErr *api.APIError
			if errors.As(err, &apiErr) && apiErr.Status == 401 {
				return fmt.Errorf("the key was rejected (401 unauthorized). double-check it's a v2 secret key, not a public SDK key")
			}
			return fmt.Errorf("verify key: %w", err)
		}
		if projectID == "" {
			projectID, err = pickProject(projects)
			if err != nil {
				return err
			}
		}
	}

	if err := store.Set(&authstore.Profile{
		Name:      loginName,
		SecretKey: loginSecretKey,
		ProjectID: projectID,
	}); err != nil {
		return err
	}

	loc := "OS keychain"
	if bypassKeychain(cmd) {
		loc = "./.revcat/config.json"
	}
	output.Success("saved profile %q to %s", loginName, loc)
	if projectID != "" {
		output.Info("  project_id: %s", projectID)
	}
	output.Info("  use it: revcat --profile %s ...", loginName)
	return nil
}

func pickProject(projects []api.Project) (string, error) {
	if len(projects) == 0 {
		return "", fmt.Errorf("the key has no project access")
	}
	if len(projects) == 1 {
		return projects[0].ID, nil
	}
	options := make([]string, len(projects))
	for i, p := range projects {
		options[i] = fmt.Sprintf("%s (%s)", p.Name, p.ID)
	}
	var idx int
	if err := survey.AskOne(&survey.Select{
		Message: "Which project?",
		Options: options,
	}, &idx); err != nil {
		return "", err
	}
	return projects[idx].ID, nil
}

func redactKey(k string) string {
	if len(k) < 8 {
		return "***"
	}
	return k[:4] + "..." + k[len(k)-4:]
}
