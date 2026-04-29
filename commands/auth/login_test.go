package auth

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
)

// newTestRoot wires up a minimal root command with the global
// --bypass-keychain flag and the auth subtree, matching the production
// shape just closely enough for runLogin to find the persistent flag.
func newTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "revcat"}
	root.PersistentFlags().Bool("bypass-keychain", false, "")
	root.AddCommand(Cmd)
	return root
}

// resetLoginFlags zeros the package-level flag-bound vars so tests
// don't leak state into each other.
func resetLoginFlags() {
	loginName = ""
	loginSecretKey = ""
	loginSecretStdin = false
	loginProjectID = ""
	loginNoVerify = false
}

// chdirTemp makes cwd a fresh temp dir and restores it on cleanup so
// that the local-file authstore writes into ./.revcat/config.json
// inside an isolated directory.
func chdirTemp(t *testing.T) string {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
	return dir
}

func TestLoginSecretKeyStdin(t *testing.T) {
	resetLoginFlags()
	dir := chdirTemp(t)

	root := newTestRoot()
	root.SetArgs([]string{
		"auth", "login",
		"--bypass-keychain",
		"--name", "test-stdin",
		"--secret-key-stdin",
		"--no-verify",
	})
	root.SetIn(strings.NewReader("sk_test_abcdef\n"))
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	// Confirm the profile landed in the local store with the trimmed key.
	store, err := authstore.Open(true)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	got, err := store.Get("test-stdin")
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if got.SecretKey != "sk_test_abcdef" {
		t.Fatalf("secret key: want %q, got %q", "sk_test_abcdef", got.SecretKey)
	}
	if _, err := os.Stat(filepath.Join(dir, ".revcat", "config.json")); err != nil {
		t.Fatalf("expected config.json: %v", err)
	}
}

func TestLoginSecretKeyAndStdinMutuallyExclusive(t *testing.T) {
	resetLoginFlags()
	chdirTemp(t)

	root := newTestRoot()
	root.SetArgs([]string{
		"auth", "login",
		"--bypass-keychain",
		"--name", "test-both",
		"--secret-key", "sk_inline",
		"--secret-key-stdin",
		"--no-verify",
	})
	root.SetIn(strings.NewReader("sk_from_stdin\n"))
	// Silence cobra's default error printing.
	root.SilenceUsage = true
	root.SilenceErrors = true
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when both --secret-key and --secret-key-stdin are passed")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error %q does not mention mutual exclusion", err.Error())
	}
}

func TestLoginSecretKeyStdinEmpty(t *testing.T) {
	resetLoginFlags()
	chdirTemp(t)

	root := newTestRoot()
	root.SetArgs([]string{
		"auth", "login",
		"--bypass-keychain",
		"--name", "test-empty",
		"--secret-key-stdin",
		"--no-verify",
	})
	root.SetIn(strings.NewReader("   \n"))
	root.SilenceUsage = true
	root.SilenceErrors = true
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for empty stdin")
	}
	if !strings.Contains(err.Error(), "stdin was empty") {
		t.Fatalf("error %q does not mention empty stdin", err.Error())
	}
}
