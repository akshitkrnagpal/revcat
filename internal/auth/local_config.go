package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// LocalConfigPath is where revcat init writes the project-local credential
// file. Walked up from cwd, gitignored. Contains a single OAuth credential
// blob plus the project_id and optional apps the user picked at init time.
const LocalConfigPath = ".revcat/config.json"

// LocalConfig is the on-disk shape of ./.revcat/config.json.
//
// Distinct from the global file shape (which is a profiles map) because
// a project directory has exactly one credential and one project_id.
// Apps are advisory metadata for app-scoped commands.
type LocalConfig struct {
	ProjectID string    `json:"project_id"`
	Apps      []LocalApp `json:"apps,omitempty"`
	Profile   Profile   `json:"profile"`

	// Path is the absolute path the config was loaded from. Empty
	// when the struct was constructed in memory (e.g. by revcat init
	// before the first save).
	Path string `json:"-"`
}

// LocalApp records one app id/name picked at init time. Mirrors the
// committed revcat.toml [[apps]] block but lives in the gitignored half
// so an agent can resolve apps without needing the toml.
type LocalApp struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// ErrNoLocalConfig is returned by LoadLocal when no .revcat/config.json
// exists walking up from cwd. Not a hard error - callers fall back to
// global creds.
var ErrNoLocalConfig = errors.New("no .revcat/config.json in this directory tree")

// LoadLocal walks up from startDir looking for .revcat/config.json.
// Returns ErrNoLocalConfig when no file is found.
func LoadLocal(startDir string) (*LocalConfig, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, err
	}
	for {
		path := filepath.Join(dir, LocalConfigPath)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return readLocalFile(path)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ErrNoLocalConfig
		}
		dir = parent
	}
}

// LoadLocalFromCwd is LoadLocal(os.Getwd()).
func LoadLocalFromCwd() (*LocalConfig, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return LoadLocal(cwd)
}

func readLocalFile(path string) (*LocalConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg LocalConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if cfg.Profile.AccessToken == "" && cfg.Profile.RefreshToken == "" {
		return nil, fmt.Errorf("%s: profile block is empty; rerun `revcat init`", path)
	}
	cfg.Path = path
	return &cfg, nil
}

// SaveLocal writes cfg to path atomically with mode 0600. Creates the
// .revcat/ directory if needed.
func SaveLocal(path string, cfg *LocalConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return atomicWriteJSON(path, cfg)
}

// EnsureGitignored appends ".revcat/" to .gitignore in dir if not
// already present. Idempotent. Creates .gitignore if missing.
//
// We don't try to detect "is this a git repo" - if .gitignore is
// useless (e.g. user is not in a git tree) the file just sits there
// without effect, which is harmless. Init prints a line confirming the
// edit so it's never a surprise.
func EnsureGitignored(dir string) (added bool, err error) {
	gi := filepath.Join(dir, ".gitignore")
	const entry = ".revcat/\n"

	existing, err := os.ReadFile(gi)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	if alreadyContains(existing, ".revcat/") {
		return false, nil
	}

	body := string(existing)
	if len(body) > 0 && body[len(body)-1] != '\n' {
		body += "\n"
	}
	body += entry

	if err := os.WriteFile(gi, []byte(body), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func alreadyContains(haystack []byte, needle string) bool {
	// Match whole lines so we don't false-positive on a comment or
	// a longer pattern. Cheap line scan.
	start := 0
	for i := 0; i <= len(haystack); i++ {
		if i == len(haystack) || haystack[i] == '\n' {
			line := string(haystack[start:i])
			line = trimSpace(line)
			if line == needle || line == needle+"/" || line == "/"+needle {
				return true
			}
			// Tolerate trailing slash variations.
			if line == ".revcat" {
				return true
			}
			start = i + 1
		}
	}
	return false
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
