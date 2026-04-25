package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// localStore persists profiles to ./.revcat/config.json. Used when
// --bypass-keychain is set or in containers where no keychain exists.
//
// File format is a flat map keyed by profile name. Auto-creates a
// .gitignore in the parent directory on first write.
type localStore struct {
	path string
}

const localFileName = ".revcat/config.json"

func openLocal() (Store, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &localStore{path: filepath.Join(wd, localFileName)}, nil
}

type localFile struct {
	Profiles map[string]Profile `json:"profiles"`
}

func (s *localStore) load() (*localFile, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &localFile{Profiles: map[string]Profile{}}, nil
		}
		return nil, err
	}
	var lf localFile
	if err := json.Unmarshal(b, &lf); err != nil {
		return nil, fmt.Errorf("parse %s: %w", s.path, err)
	}
	if lf.Profiles == nil {
		lf.Profiles = map[string]Profile{}
	}
	return &lf, nil
}

func (s *localStore) save(lf *localFile) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	if err := writeGitignoreOnce(filepath.Dir(s.path)); err != nil {
		return err
	}
	b, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o600)
}

func writeGitignoreOnce(dir string) error {
	gi := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gi); err == nil {
		return nil
	}
	return os.WriteFile(gi, []byte("config.json\n"), 0o644)
}

func (s *localStore) Get(name string) (*Profile, error) {
	lf, err := s.load()
	if err != nil {
		return nil, err
	}
	p, ok := lf.Profiles[name]
	if !ok {
		return nil, ErrNoProfile
	}
	return &p, nil
}

func (s *localStore) Set(p *Profile) error {
	lf, err := s.load()
	if err != nil {
		return err
	}
	lf.Profiles[p.Name] = *p
	return s.save(lf)
}

func (s *localStore) Delete(name string) error {
	lf, err := s.load()
	if err != nil {
		return err
	}
	if _, ok := lf.Profiles[name]; !ok {
		return ErrNoProfile
	}
	delete(lf.Profiles, name)
	return s.save(lf)
}

func (s *localStore) List() ([]string, error) {
	lf, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(lf.Profiles))
	for k := range lf.Profiles {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}
