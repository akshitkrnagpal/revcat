// Package output renders command results to stdout in the format the
// caller wants - table when a human is watching, JSON when a script is.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type Format int

const (
	FormatAuto Format = iota
	FormatTable
	FormatJSON
	FormatCSV
	FormatMarkdown
)

// Options is what commands/root passes us during PersistentPreRun.
type Options struct {
	Format  string
	Pretty  bool
	NoColor bool
	Quiet   bool
	Verbose bool
}

var current = struct {
	format  Format
	pretty  bool
	noColor bool
	quiet   bool
	verbose bool
	stdout  io.Writer
	stderr  io.Writer
}{
	stdout: os.Stdout,
	stderr: os.Stderr,
}

// Configure resolves auto-format and stores the rendering options.
func Configure(opts Options) {
	current.format = parseFormat(opts.Format)
	current.pretty = opts.Pretty
	current.noColor = opts.NoColor || os.Getenv("NO_COLOR") != ""
	current.quiet = opts.Quiet
	current.verbose = opts.Verbose
	if env := os.Getenv("REVCAT_DEFAULT_OUTPUT"); env != "" && opts.Format == "" {
		current.format = parseFormat(env)
	}
	if current.noColor {
		lipgloss.SetColorProfile(0)
	}
}

func parseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "table":
		return FormatTable
	case "json":
		return FormatJSON
	case "csv":
		return FormatCSV
	case "markdown", "md":
		return FormatMarkdown
	}
	return FormatAuto
}

// Resolved returns the format to actually use, expanding "auto" by checking
// whether stdout is a TTY.
func Resolved() Format {
	if current.format != FormatAuto {
		return current.format
	}
	if isTerminal(os.Stdout) {
		return FormatTable
	}
	return FormatJSON
}

// IsJSON reports whether the active output is JSON-shaped.
func IsJSON() bool { return Resolved() == FormatJSON }

func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// JSON marshals v to stdout as JSON. Honors --pretty.
func JSON(v any) error {
	enc := json.NewEncoder(current.stdout)
	if current.pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}

// Info writes a non-essential message to stderr unless --quiet.
func Info(format string, args ...any) {
	if current.quiet {
		return
	}
	fmt.Fprintf(current.stderr, format+"\n", args...)
}

// Success writes a green checkmark line to stderr unless --quiet.
func Success(format string, args ...any) {
	if current.quiet {
		return
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	fmt.Fprintln(current.stderr, style.Render("✓ ")+fmt.Sprintf(format, args...))
}

// Warn writes a yellow warning line to stderr.
func Warn(format string, args ...any) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	fmt.Fprintln(current.stderr, style.Render("! ")+fmt.Sprintf(format, args...))
}

// Errorln writes a red error line to stderr.
func Errorln(format string, args ...any) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	fmt.Fprintln(current.stderr, style.Render("✗ ")+fmt.Sprintf(format, args...))
}
