package auth

import (
	"errors"
	"fmt"
	"os"
)

// envRefreshToken is the CI / sandbox / agent escape hatch. When set,
// resolution short-circuits and returns a virtual profile carrying just
// the refresh token; callers refresh it on first API call to get a
// usable access token. Pair with REVCAT_PROJECT_ID to skip both
// keychain and walk-up file lookup.
const envRefreshToken = "REVCAT_REFRESH_TOKEN"

// envClientIDForToken lets the env hatch override the OAuth client_id
// the refresh call uses. Defaults to the embedded public client.
const envClientIDForToken = "REVCAT_OAUTH_CLIENT_ID"

// Source records where a Resolved credential came from. Surfaced by
// `revcat auth status` so the user can debug "why am I authed as that?".
type Source string

const (
	SourceLocal       Source = "local"
	SourceKeychain    Source = "keychain"
	SourceGlobalFile  Source = "file"
	SourceEnv         Source = "env"
	SourceUnknown     Source = "unknown"
)

// Resolved is the unified output of credential resolution: which
// profile applies, what its project_id is (when known here), and where
// it came from.
type Resolved struct {
	Profile   *Profile
	ProjectID string
	Apps      []LocalApp
	Source    Source

	// Path is set for SourceLocal (the .revcat/config.json path) and
	// SourceGlobalFile (~/.revcat/config.json). Empty otherwise.
	Path string

	// Local is the loaded LocalConfig when Source == SourceLocal.
	// Useful for callers that want to write the profile back after
	// refresh.
	Local *LocalConfig
}

// ResolveOptions tweaks resolution. Bypass forces the file backend.
// FlagProfile is the value of the global --profile flag (used only
// when no local config or env hatch wins).
type ResolveOptions struct {
	Bypass      bool
	FlagProfile string
	Cwd         string
}

// Resolve walks the precedence chain and returns the active credential.
//
//	1. REVCAT_REFRESH_TOKEN env (synthesizes a Profile, no on-disk state)
//	2. ./.revcat/config.json walked up from cwd
//	3. Global keychain or ~/.revcat/config.json (file backend) for the
//	   active profile name
func Resolve(opts ResolveOptions) (*Resolved, error) {
	if v := os.Getenv(envRefreshToken); v != "" {
		return &Resolved{
			Profile: &Profile{
				Name:         "$" + envRefreshToken,
				RefreshToken: v,
				ClientID:     os.Getenv(envClientIDForToken),
			},
			ProjectID: os.Getenv("REVCAT_PROJECT_ID"),
			Source:    SourceEnv,
		}, nil
	}

	cwd := opts.Cwd
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getwd: %w", err)
		}
	}

	if local, err := LoadLocal(cwd); err == nil {
		return &Resolved{
			Profile:   &local.Profile,
			ProjectID: local.ProjectID,
			Apps:      local.Apps,
			Source:    SourceLocal,
			Path:      local.Path,
			Local:     local,
		}, nil
	} else if !errors.Is(err, ErrNoLocalConfig) {
		return nil, err
	}

	store, err := OpenGlobal(opts.Bypass)
	if err != nil {
		return nil, err
	}
	name := ResolveProfileName(opts.FlagProfile)
	prof, err := store.Get(name)
	if err != nil {
		return nil, err
	}

	source := SourceKeychain
	path := ""
	if opts.Bypass || os.Getenv(envBypassKeychain) == "1" {
		source = SourceGlobalFile
		if home, herr := os.UserHomeDir(); herr == nil {
			path = home + "/" + globalFileName
		}
	}
	return &Resolved{
		Profile:   prof,
		ProjectID: "",
		Source:    source,
		Path:      path,
	}, nil
}
