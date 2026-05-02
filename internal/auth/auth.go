// Package auth handles OAuth credential storage and resolution.
//
// Two storage tiers:
//
//   - Global (~/.revcat/config.json, mode 0600): created by
//     `revcat auth login`. Holds OAuth tokens keyed by profile name.
//     Multi-account is supported here via --profile.
//
//   - Project-local (./.revcat/config.json, gitignored, mode 0600):
//     created by `revcat init` in a repo. Single credential blob plus
//     project_id and optional apps. Walked up from cwd. When present,
//     overrides the global profile so an agent or sandbox in the
//     directory can run without touching ~/.
//
// Credential resolution (Resolve below):
//
//  1. REVCAT_REFRESH_TOKEN env (CI / sandbox hatch).
//  2. ./.revcat/config.json walked up from cwd.
//  3. ~/.revcat/config.json for the active profile.
//
// Schema rule: as of v0.4 the only credential type is OAuth. Profiles
// stored under the v0.3 secret-key shape error on read with a clear
// hint to rerun `revcat auth login`.
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// envProfile picks the active profile when --profile is not set.
const envProfile = "REVCAT_PROFILE"

const defaultProfile = "default"

// Profile is one set of OAuth credentials persisted under a name.
type Profile struct {
	Name         string `json:"name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at_ms"`
	Scope        string `json:"scope,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
}

// Token is the bearer token to send to the API. Equal to AccessToken
// for valid profiles; the OAuthTokenSource refreshes ahead of expiry.
func (p *Profile) Token() string { return p.AccessToken }

// NeedsRefresh is true when the access token is within skew of expiry
// (or already expired or unknown).
func (p *Profile) NeedsRefresh(skew time.Duration) bool {
	if p.ExpiresAt == 0 {
		return true
	}
	return time.Now().Add(skew).After(time.UnixMilli(p.ExpiresAt))
}

// ErrNoProfile is returned when the requested global profile does not
// exist.
var ErrNoProfile = errors.New("no profile found; run `revcat auth login`")

// ErrLegacyProfile is returned when the on-disk shape is from v0.3
// (secret-key auth). The user must rerun login under v0.4+ OAuth.
var ErrLegacyProfile = errors.New("this profile was created under v0.3 secret-key auth, which was removed in v0.4. run `revcat auth login` to reauth via OAuth")

// GlobalStore is the credential persistence interface for the global
// tier. Always backed by ~/.revcat/config.json since v0.6 (the
// passphrase-encrypted keyring backend was dropped because the
// shipped binary's CGO_ENABLED=0 build couldn't reach the real OS
// keychain anyway, so the encryption added friction without real
// security).
type GlobalStore interface {
	Get(name string) (*Profile, error)
	Set(p *Profile) error
	Delete(name string) error
	List() ([]string, error)
}

// OpenGlobal returns the global store. Always the file backend at
// ~/.revcat/config.json (mode 0600).
func OpenGlobal() (GlobalStore, error) {
	return openGlobalFile()
}

// ResolveProfileName picks which global profile to use. Precedence:
//
//	flagProfile > REVCAT_PROFILE > ~/.revcat/active > "default"
func ResolveProfileName(flagProfile string) string {
	if flagProfile != "" {
		return flagProfile
	}
	if v := os.Getenv(envProfile); v != "" {
		return v
	}
	if active, _ := GetActive(); active != "" {
		return active
	}
	return defaultProfile
}

// decodeProfile parses a profile blob and rejects v0.3 secret-key
// shapes with ErrLegacyProfile so the user gets a clear migration hint.
func decodeProfile(data []byte, name string) (*Profile, error) {
	// Detect legacy shape: a populated secret_key field (or the
	// presence of "secret_key" at all where access_token is empty).
	var probe struct {
		SecretKey   string `json:"secret_key"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("decode profile %q: %w", name, err)
	}
	if probe.SecretKey != "" && probe.AccessToken == "" {
		return nil, ErrLegacyProfile
	}
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("decode profile %q: %w", name, err)
	}
	if p.AccessToken == "" {
		return nil, ErrLegacyProfile
	}
	if p.Name == "" {
		p.Name = name
	}
	return &p, nil
}
