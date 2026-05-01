package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// globalFileStore is the GlobalStore backed by ~/.revcat/config.json.
// Used when --bypass-keychain is set or REVCAT_BYPASS_KEYCHAIN=1, in
// containers / Linux without secret-service / CI without a keyring.
//
// File format: a flat map keyed by profile name, mode 0600 to keep
// other users on the box from reading it. Atomic writes via
// atomicWriteJSON so a Ctrl-C mid-write doesn't corrupt the file.
type globalFileStore struct {
	path string
}

// globalFileName is the path under HOME the file backend writes to.
// Pre-v0.4 it lived at the cwd, which made bypass-keychain depend on
// where you ran revcat from. Now it's HOME-anchored, mirroring git's
// ~/.gitconfig pattern.
const globalFileName = ".revcat/config.json"

func openGlobalFile() (GlobalStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("locate home dir: %w", err)
	}
	return &globalFileStore{path: filepath.Join(home, globalFileName)}, nil
}

type globalFileShape struct {
	Profiles map[string]json.RawMessage `json:"profiles"`
}

func (s *globalFileStore) load() (*globalFileShape, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &globalFileShape{Profiles: map[string]json.RawMessage{}}, nil
		}
		return nil, fmt.Errorf("read %s: %w", s.path, err)
	}
	var f globalFileShape
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", s.path, err)
	}
	if f.Profiles == nil {
		f.Profiles = map[string]json.RawMessage{}
	}
	return &f, nil
}

func (s *globalFileStore) save(f *globalFileShape) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	return atomicWriteJSON(s.path, f)
}

func (s *globalFileStore) Get(name string) (*Profile, error) {
	f, err := s.load()
	if err != nil {
		return nil, err
	}
	raw, ok := f.Profiles[name]
	if !ok {
		return nil, ErrNoProfile
	}
	return decodeProfile(raw, name)
}

func (s *globalFileStore) Set(p *Profile) error {
	f, err := s.load()
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(p)
	if err != nil {
		return err
	}
	f.Profiles[p.Name] = encoded
	return s.save(f)
}

func (s *globalFileStore) Delete(name string) error {
	f, err := s.load()
	if err != nil {
		return err
	}
	if _, ok := f.Profiles[name]; !ok {
		return ErrNoProfile
	}
	delete(f.Profiles, name)
	return s.save(f)
}

func (s *globalFileStore) List() ([]string, error) {
	f, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(f.Profiles))
	for k := range f.Profiles {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}
