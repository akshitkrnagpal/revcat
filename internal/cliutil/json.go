package cliutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// MaxJSONSize caps the size of a JSON payload accepted from --file or
// stdin. Set deliberately low (4 MiB) since RC payloads (paywall configs,
// CRUD bodies) are far smaller in practice. A larger input is almost
// always a mistake or an attempt to OOM the process.
const MaxJSONSize = 4 << 20

// tooLargeErr formats the size-cap error so both the stdin and --file
// branches produce identical wording.
func tooLargeErr(size int64) error {
	return fmt.Errorf("input too large: file is %d bytes, max is 4 MiB. Pipe via stdin if you really need more (rare for paywall configs).", size)
}

// LoadJSON reads a JSON file into a generic map. Path "-" reads stdin.
// Used by every CRUD command that takes --file. Both branches enforce
// MaxJSONSize so a giant file can't OOM the process.
func LoadJSON(path string) (map[string]any, error) {
	if path == "" {
		return nil, errors.New("--file is required")
	}
	b, err := readCapped(path)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return out, nil
}

// ReadCappedFile reads a file (or stdin if path == "-") into memory,
// rejecting anything larger than MaxJSONSize. Exported so other packages
// (e.g. commands/publish) can apply the same cap to user-provided paths.
func ReadCappedFile(path string) ([]byte, error) {
	return readCapped(path)
}

func readCapped(path string) ([]byte, error) {
	if path == "-" {
		b, err := io.ReadAll(io.LimitReader(os.Stdin, MaxJSONSize+1))
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		if int64(len(b)) > MaxJSONSize {
			return nil, tooLargeErr(int64(len(b)))
		}
		return b, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	defer f.Close()
	// Stat first so we can give a precise byte count in the error;
	// fall back to the LimitReader path if Stat fails or reports 0
	// (pipes, /dev/stdin via path, etc).
	if fi, statErr := f.Stat(); statErr == nil && fi.Mode().IsRegular() && fi.Size() > MaxJSONSize {
		return nil, tooLargeErr(fi.Size())
	}
	b, err := io.ReadAll(io.LimitReader(f, MaxJSONSize+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if int64(len(b)) > MaxJSONSize {
		return nil, tooLargeErr(int64(len(b)))
	}
	return b, nil
}
