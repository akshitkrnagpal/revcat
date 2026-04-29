package auth

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func TestAtomicWriteJSON_HappyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	payload := localFile{Profiles: map[string]Profile{
		"default": {Name: "default", SecretKey: "sk_live_abc", ProjectID: "proj_1"},
	}}
	if err := atomicWriteJSON(path, &payload); err != nil {
		t.Fatalf("atomicWriteJSON: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if runtime.GOOS != "windows" {
		if got := info.Mode().Perm(); got != 0o600 {
			t.Errorf("file mode = %o, want 0600", got)
		}
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	// Cheap content check: payload should round-trip through the JSON we wrote.
	if !contains(b, []byte(`"sk_live_abc"`)) || !contains(b, []byte(`"proj_1"`)) {
		t.Errorf("content missing expected fields: %s", b)
	}
}

func TestAtomicWriteJSON_NoLeftoverTempfileOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := atomicWriteJSON(path, map[string]string{"k": "v"}); err != nil {
		t.Fatalf("atomicWriteJSON: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "config.json" {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("expected only config.json, got %v", names)
	}
}

func TestAtomicWriteWith_FailureLeavesDestinationUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Pre-populate with a known good file.
	original := []byte(`{"profiles":{"default":{"name":"default","secret_key":"original"}}}` + "\n")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Simulate a writer that errors after a few bytes.
	failErr := errors.New("simulated mid-write failure")
	werr := atomicWriteWith(path, 0o600, func(w io.Writer) error {
		if _, err := w.Write([]byte("garbage-prefix")); err != nil {
			return err
		}
		return failErr
	})
	if !errors.Is(werr, failErr) {
		t.Fatalf("expected wrapped failErr, got %v", werr)
	}

	// Destination must still contain the original bytes.
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(original) {
		t.Errorf("destination mutated.\n got: %s\nwant: %s", got, original)
	}

	// And no tempfile should be left in dir.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 1 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("expected only config.json after failure, got %v", names)
	}
}

func TestAtomicWriteFile_ConcurrentWritesNoCorruption(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	a := []byte(`{"profiles":{"a":{"name":"a","secret_key":"AAA"}}}` + "\n")
	b := []byte(`{"profiles":{"b":{"name":"b","secret_key":"BBB"}}}` + "\n")

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			if err := atomicWriteFile(path, a, 0o600); err != nil {
				t.Errorf("writer A: %v", err)
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			if err := atomicWriteFile(path, b, 0o600); err != nil {
				t.Errorf("writer B: %v", err)
				return
			}
		}
	}()
	wg.Wait()

	final, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	// File must contain exactly one of the two contents, never an interleaved
	// fragment.
	if string(final) != string(a) && string(final) != string(b) {
		t.Errorf("file content is neither A nor B: %s", final)
	}
}

func TestAtomicWriteJSON_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := atomicWriteJSON(path, map[string]int{"v": 1}); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := atomicWriteJSON(path, map[string]int{"v": 2}); err != nil {
		t.Fatalf("second write: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !contains(got, []byte(`"v": 2`)) {
		t.Errorf("expected v=2 in final file, got %s", got)
	}
}

func contains(haystack, needle []byte) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
