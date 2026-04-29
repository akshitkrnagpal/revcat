package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// atomicWriteJSON marshals v as indented JSON and writes it to path
// atomically: a tempfile is created in the destination directory, mode is
// fixed to 0600 BEFORE any bytes are written, the bytes are flushed to
// disk, and then the tempfile is renamed over path.
//
// On any failure the tempfile is removed so we don't litter the directory.
// A successful rename preserves the tempfile's 0600 mode on the final file.
func atomicWriteJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(path, b, 0o600)
}

// atomicWriteFile writes data to path via tempfile + rename. Mode is set on
// the tempfile before writing, so the final file has the requested mode.
//
// writerFn lets tests inject a writer that fails mid-stream.
func atomicWriteFile(path string, data []byte, mode os.FileMode) error {
	return atomicWriteWith(path, mode, func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	})
}

// atomicWriteWith is the underlying primitive: it creates a tempfile in
// dir(path), chmods to mode, calls writeFn with the file as an io.Writer,
// fsyncs, closes, and renames. On any error the tempfile is removed.
func atomicWriteWith(path string, mode os.FileMode, writeFn func(io.Writer) error) (err error) {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".revcat-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tempPath := f.Name()

	// Always try to remove the tempfile. After a successful rename the
	// original tempPath no longer exists, so Remove is a harmless no-op.
	defer func() {
		_ = os.Remove(tempPath)
	}()

	// Set restrictive mode BEFORE writing so the bytes never sit on disk
	// world-readable. On platforms where Chmod is a no-op (Windows) this
	// is best-effort; the rename below still gives an atomic swap.
	if chmodErr := os.Chmod(tempPath, mode); chmodErr != nil {
		_ = f.Close()
		return fmt.Errorf("chmod temp file: %w", chmodErr)
	}

	if writeErr := writeFn(f); writeErr != nil {
		_ = f.Close()
		return writeErr
	}

	if syncErr := f.Sync(); syncErr != nil {
		_ = f.Close()
		return fmt.Errorf("sync temp file: %w", syncErr)
	}

	if closeErr := f.Close(); closeErr != nil {
		return fmt.Errorf("close temp file: %w", closeErr)
	}

	if renameErr := os.Rename(tempPath, path); renameErr != nil {
		return fmt.Errorf("rename temp file: %w", renameErr)
	}
	return nil
}
