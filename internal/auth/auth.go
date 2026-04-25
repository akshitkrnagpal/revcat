// Package auth handles credential storage and profile resolution.
//
// Two backends:
//   - keychain (default): 99designs/keyring, OS-native secure storage
//   - local file: ./.revcat/config.json, used when --bypass-keychain is set
//     or REVCAT_BYPASS_KEYCHAIN=1
//
// A "profile" is a named credential set: {name, secret_key, project_id?}.
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/99designs/keyring"
)

// service is the keychain service identifier. All revcat profiles live under
// this service; the keychain account name is the profile name.
const service = "revcat"

// envKeyOverride lets a script bypass profiles entirely by passing the
// secret key through an env var. Highest precedence in Resolve().
const envKeyOverride = "REVCAT_API_KEY"

// envBypassKeychain forces the local-file backend even without --bypass-keychain.
const envBypassKeychain = "REVCAT_BYPASS_KEYCHAIN"

// envProfile picks the active profile when --profile is not set.
const envProfile = "REVCAT_PROFILE"

// envProjectID lets a script override the stored project id.
const envProjectID = "REVCAT_PROJECT_ID"

const defaultProfile = "default"

// Profile is the credential set we persist per name.
type Profile struct {
	Name      string `json:"name"`
	SecretKey string `json:"secret_key"`
	ProjectID string `json:"project_id,omitempty"`
}

// ErrNoProfile is returned when the requested profile doesn't exist.
var ErrNoProfile = errors.New("no profile found; run `revcat auth login`")

// Store is the credential persistence interface. Two implementations
// (keychain, localFile) selected at runtime.
type Store interface {
	Get(name string) (*Profile, error)
	Set(p *Profile) error
	Delete(name string) error
	List() ([]string, error)
}

// Open returns the right store for the current process. bypass=true forces
// the local-file backend.
func Open(bypass bool) (Store, error) {
	if bypass || os.Getenv(envBypassKeychain) == "1" {
		return openLocal()
	}
	return openKeychain()
}

// Resolve returns the active profile per the precedence order:
//  1. REVCAT_API_KEY env var (synthesizes an unnamed profile)
//  2. --profile flag
//  3. REVCAT_PROFILE env var
//  4. ~/.revcat/active (set by `revcat auth use`)
//  5. "default"
func Resolve(store Store, flagProfile string) (*Profile, error) {
	if key := os.Getenv(envKeyOverride); key != "" {
		return &Profile{
			Name:      "$" + envKeyOverride,
			SecretKey: key,
			ProjectID: os.Getenv(envProjectID),
		}, nil
	}
	name := flagProfile
	if name == "" {
		name = os.Getenv(envProfile)
	}
	if name == "" {
		if active, _ := GetActive(); active != "" {
			name = active
		}
	}
	if name == "" {
		name = defaultProfile
	}
	p, err := store.Get(name)
	if err != nil {
		return nil, err
	}
	if pid := os.Getenv(envProjectID); pid != "" {
		p.ProjectID = pid
	}
	return p, nil
}

// ----- keychain backend -----

type keychainStore struct{ ring keyring.Keyring }

func openKeychain() (Store, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName:              service,
		KeychainTrustApplication: true,
		KeychainSynchronizable:   false,
		FileDir:                  "~/.revcat/keyring",
		FilePasswordFunc:         keyring.TerminalPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("open keychain: %w", err)
	}
	return &keychainStore{ring: ring}, nil
}

func (s *keychainStore) Get(name string) (*Profile, error) {
	item, err := s.ring.Get(name)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return nil, ErrNoProfile
		}
		return nil, fmt.Errorf("read keychain: %w", err)
	}
	var p Profile
	if err := json.Unmarshal(item.Data, &p); err != nil {
		return nil, fmt.Errorf("decode profile %q: %w", name, err)
	}
	return &p, nil
}

func (s *keychainStore) Set(p *Profile) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return s.ring.Set(keyring.Item{
		Key:         p.Name,
		Data:        data,
		Label:       "revcat: " + p.Name,
		Description: "RevenueCat secret key",
	})
}

func (s *keychainStore) Delete(name string) error {
	if err := s.ring.Remove(name); err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return ErrNoProfile
		}
		return err
	}
	return nil
}

func (s *keychainStore) List() ([]string, error) {
	keys, err := s.ring.Keys()
	if err != nil {
		return nil, err
	}
	return keys, nil
}
