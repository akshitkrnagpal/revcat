package cliutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

const maxJSONInput = 4 * 1024 * 1024

// LoadJSON reads a JSON file into a generic map. Path "-" reads stdin.
// Used by every CRUD command that takes --file.
func LoadJSON(path string) (map[string]any, error) {
	if path == "" {
		return nil, errors.New("--file is required")
	}
	var b []byte
	var err error
	if path == "-" {
		b, err = io.ReadAll(io.LimitReader(os.Stdin, maxJSONInput+1))
		if err == nil && len(b) > maxJSONInput {
			return nil, errors.New("input too large")
		}
	} else {
		b, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return out, nil
}
