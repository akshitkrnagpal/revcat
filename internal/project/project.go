// Package project loads per-repo revcat configuration. The config lives
// in revcat.toml at the repo root and is committed alongside the source,
// the same way Terraform pins its providers in main.tf.
//
// Resolution: walked up from the current working directory until a
// revcat.toml is found (like .git or go.mod). Missing file is not an
// error - it just means "no project context", and callers fall back to
// flag/env/legacy-profile.
package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// FileName is the per-repo config file. Walked up from cwd.
const FileName = "revcat.toml"

// Config is the parsed revcat.toml. Apps is optional - useful for store
// bridges where multiple platform bundles share a project.
type Config struct {
	ProjectID string `toml:"project_id"`
	Apps      []App  `toml:"apps"`

	// Path is the absolute path to the file this Config was loaded from.
	// Empty if no file was found (default Config{}).
	Path string `toml:"-"`
}

// App is one platform bundle declared in revcat.toml. Optional metadata
// only; commands resolve apps by id.
type App struct {
	ID   string `toml:"id"`
	Name string `toml:"name,omitempty"`
}

// ErrNotFound is returned by Load when no revcat.toml is found walking
// up from startDir to the filesystem root.
var ErrNotFound = errors.New("no revcat.toml found")

// Load walks up from startDir looking for revcat.toml. Returns
// ErrNotFound when no file exists; other errors (parse, IO) bubble up
// unwrapped.
func Load(startDir string) (*Config, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, err
	}
	for {
		path := filepath.Join(dir, FileName)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return readFile(path)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ErrNotFound
		}
		dir = parent
	}
}

// LoadFromCwd is Load(os.Getwd()).
func LoadFromCwd() (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return Load(cwd)
}

func readFile(path string) (*Config, error) {
	var cfg Config
	meta, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		// Tolerate forward-compatible keys (env blocks etc) - the
		// decoder ignored them, we just don't surface a warning.
		_ = undecoded
	}
	cfg.Path = path
	return &cfg, nil
}

// Save writes cfg to path atomically. Used by `revcat init`.
func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".revcat-toml-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	enc := toml.NewEncoder(tmp)
	if err := enc.Encode(toCanonical(cfg)); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// toCanonical produces a stable, comment-free shape for the encoder so
// the file we write looks the same across runs (Path is a transient
// loaded-from field, not on-disk data).
type canonical struct {
	ProjectID string `toml:"project_id"`
	Apps      []App  `toml:"apps,omitempty"`
}

func toCanonical(cfg *Config) canonical {
	return canonical{ProjectID: cfg.ProjectID, Apps: cfg.Apps}
}
