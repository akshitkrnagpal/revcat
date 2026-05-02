package commands

// Drift detector. Walks the cobra tree to build the set of real
// command paths, then scans every Markdown file in the repo for
// `revcat <command path>` mentions inside code spans (inline
// backticks or fenced code blocks). Each mention must resolve to a
// real command path — otherwise it's drift.
//
// Run locally:
//
//	go test ./commands -run TestDocsDontReferenceMissingCommands -v
//
// Or via the make target:
//
//	make drift-check
//
// Coverage:
//
//   - Catches the common drift class: a doc / skill names a subcommand
//     that doesn't exist (e.g. v0.3-era `revcat events tail`), or a
//     previously-removed subcommand still on a surface map (e.g.
//     `vc-balance`).
//
// Out of scope:
//
//   - Flag-level drift (placeholder shapes like `--start YYYY-MM-DD`
//     are hard to disambiguate from real flags). Skip for v1.
//   - Prose claims (e.g. "needs a partner-tier key") that aren't
//     command-shaped. Only humans can spot those.
//
// Why only code spans: scanning running prose flags any sentence that
// starts with "revcat" (e.g. "revcat doesn't expose ..."). Real
// command examples live in backticks or fenced blocks; prose mentions
// don't.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// allowlistedFirstToken handles command-shaped tokens that are
// instructions to the reader, not literal commands. They start with
// `revcat ` and the first token is a placeholder.
var allowlistedFirstToken = map[string]bool{
	"<command>":  true,
	"<group>":    true,
	"<resource>": true,
	"<id>":       true,
	"<cmd>":      true,
	"--profile":  true, // e.g. "revcat --profile work auth status"
	"--bypass-keychain": true,
}

// inlineRe matches an inline code span containing `revcat ...`.
var inlineRe = regexp.MustCompile("`revcat ([^`]+)`")

// fenceRe captures the body of a fenced code block.
var fenceRe = regexp.MustCompile("(?s)```[A-Za-z]*\n(.*?)\n```")

// fenceLineRe matches a `revcat <args>` invocation as it appears on a
// line inside a fenced block. The optional leading `$ ` is the demo
// prompt prefix, also accept a trailing-space tab.
var fenceLineRe = regexp.MustCompile(`(?m)^(?:\$\s+)?revcat (.+)$`)

// commandPathTokens extracts the command-path tokens from the tail of
// a `revcat ` invocation. Stops at the first non-command token: a
// flag (--), a placeholder (<), an option-list ([), an uppercase
// (likely an env or arg), a comment (#), or end of input.
var commandPathTokens = regexp.MustCompile(`^([a-z][a-z0-9-]*(?:\s+[a-z][a-z0-9-]*)*)`)

func TestDocsDontReferenceMissingCommands(t *testing.T) {
	root := RootCmd()
	// Cobra adds `completion` lazily on Execute(). Force-init so its
	// subcommands (bash / zsh / fish / powershell) are visible to
	// the walker.
	root.InitDefaultCompletionCmd()
	known := collectCommandPaths(root)
	known["help"] = true
	known["version"] = true

	repoRoot := mustRepoRoot(t)
	files := mustFindMarkdown(t, repoRoot)

	var failures []string
	for _, file := range files {
		if strings.HasSuffix(file, "/reference/cli.md") {
			// Auto-generated tier-2 doc; pulls from cobra by
			// definition.
			continue
		}
		if filepath.Base(file) == "CHANGELOG.md" {
			// Historical record - mentions of removed commands are
			// expected (e.g. "removed `revcat events list`").
			continue
		}
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(body)

		// Inline code spans: `revcat foo bar`
		for _, m := range inlineRe.FindAllStringSubmatch(text, -1) {
			path := extractCommandPath(m[1])
			if path == "" || allowlistedFirstToken[firstToken(path)] {
				continue
			}
			if isKnownPathOrPrefix(known, path) {
				continue
			}
			rel, _ := filepath.Rel(repoRoot, file)
			failures = append(failures, fmt.Sprintf("%s: `revcat %s` — not a real command path", rel, path))
		}

		// Fenced code blocks: scan each line for `revcat ...`
		for _, fence := range fenceRe.FindAllStringSubmatch(text, -1) {
			for _, lm := range fenceLineRe.FindAllStringSubmatch(fence[1], -1) {
				path := extractCommandPath(lm[1])
				if path == "" || allowlistedFirstToken[firstToken(path)] {
					continue
				}
				if isKnownPathOrPrefix(known, path) {
					continue
				}
				rel, _ := filepath.Rel(repoRoot, file)
				failures = append(failures, fmt.Sprintf("%s: `revcat %s` (in code block) — not a real command path", rel, path))
			}
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		failures = dedup(failures)
		t.Errorf("found %d drift hits:\n  %s", len(failures), strings.Join(failures, "\n  "))
	}
}

// extractCommandPath returns the lowercase command path at the start
// of `tail` (the substring after `revcat `), stopping at the first
// non-path token.
func extractCommandPath(tail string) string {
	tail = strings.TrimSpace(tail)
	m := commandPathTokens.FindStringSubmatch(tail)
	if len(m) < 2 {
		// Could be `revcat <group>` or similar; surface the first token
		// for allowlist filtering.
		fields := strings.Fields(tail)
		if len(fields) == 0 {
			return ""
		}
		return fields[0]
	}
	// Normalize internal whitespace.
	return strings.Join(strings.Fields(m[1]), " ")
}

func firstToken(path string) string {
	fields := strings.Fields(path)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func collectCommandPaths(c *cobra.Command) map[string]bool {
	out := map[string]bool{}
	var walk func(cmd *cobra.Command, path []string)
	walk = func(cmd *cobra.Command, path []string) {
		if len(path) > 0 {
			out[strings.Join(path, " ")] = true
		}
		for _, sub := range cmd.Commands() {
			walk(sub, append(append([]string{}, path...), sub.Name()))
		}
		// Index aliases at this level (e.g. `subscribers` aliased to
		// `customers`).
		for _, alias := range cmd.Aliases {
			if len(path) == 0 {
				continue
			}
			aliased := append([]string{}, path[:len(path)-1]...)
			aliased = append(aliased, alias)
			out[strings.Join(aliased, " ")] = true
			// Also their subcommands.
			for _, sub := range cmd.Commands() {
				out[strings.Join(append(aliased, sub.Name()), " ")] = true
			}
		}
	}
	walk(c, nil)
	return out
}

func isKnownPathOrPrefix(known map[string]bool, path string) bool {
	if known[path] {
		return true
	}
	parts := strings.Fields(path)
	for i := len(parts) - 1; i >= 1; i-- {
		prefix := strings.Join(parts[:i], " ")
		if known[prefix] {
			return true
		}
	}
	return false
}

func mustRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find go.mod walking up from cwd")
		}
		dir = parent
	}
}

func mustFindMarkdown(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == "dist" || name == ".git" || name == ".github" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, ".mdx") {
			return nil
		}
		out = append(out, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func dedup(ss []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range ss {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
