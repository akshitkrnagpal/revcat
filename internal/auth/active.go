package auth

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// activeFile is where `revcat auth use <name>` writes the active profile
// name. Stays out of the keychain because it's not a secret and we want
// it cheap to read on every command.
//
// Lookup precedence handled in Resolve(): --profile flag > REVCAT_PROFILE
// env > activeFile > "default".
const activeFileName = "active"

func activeFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".revcat", activeFileName), nil
}

// SetActive persists the active profile name. Creates ~/.revcat if missing.
//
// Writes are atomic via tempfile + rename so a Ctrl-C mid-write or two
// concurrent `revcat auth use` invocations cannot leave a half-written
// active file.
func SetActive(name string) error {
	path, err := activeFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return atomicWriteFile(path, []byte(name+"\n"), 0o600)
}

// GetActive reads the active profile name. Returns empty string if unset
// (caller falls back to "default").
func GetActive() (string, error) {
	path, err := activeFilePath()
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
