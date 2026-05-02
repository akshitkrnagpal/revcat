// gen-group-docs regenerates the auto-generated portion of every
// docs/src/content/docs/commands/<group>.md file from the cobra tree.
//
// Each per-group page has two halves separated by an HTML comment
// marker:
//
//	[generated header: frontmatter + intro + subcommands table]
//	<!-- AUTOGEN_END -->
//	[hand-written: examples, conceptual blocks, FAQ]
//
// Running this regenerator rewrites everything ABOVE the marker from
// cobra; everything BELOW is preserved verbatim. New pages get a
// minimal hand-written tail with a "## Examples" stub.
//
// Usage (from the repo root):
//
//	go run ./scripts/gen-group-docs
//
// Or via the make target:
//
//	make docs-groups
//
// Drift discipline: the generated header includes a link to the
// canonical CLI reference at /reference/cli/, so the per-group page
// stays as a topical entrypoint rather than a competing source of
// truth for flag details.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/commands"
)

const (
	docsDir   = "docs/src/content/docs/commands"
	autogenEnd = "<!-- AUTOGEN_END -->"
)

// topLevelGroups maps cobra command names to docs filenames. The
// docs file is by convention `<command>.md`, but we map explicitly
// so we don't accidentally regenerate anything we didn't intend.
//
// Top-level commands without a per-group page (e.g. version, help,
// completion) are intentionally absent.
var topLevelGroups = []string{
	"apps",
	"audit-logs",
	"auth",
	"charts",
	"collaborators",
	"doctor",
	"entitlements",
	"init",
	"invoices",
	"metrics",
	"offerings",
	"packages",
	"paywalls",
	"products",
	"projects",
	"publish",
	"purchases",
	"subscribers",
	"subscriptions",
	"virtual-currencies",
	"webhooks",
}

func main() {
	root := commands.RootCmd()
	root.InitDefaultCompletionCmd()

	byName := map[string]*cobra.Command{}
	for _, c := range root.Commands() {
		byName[c.Name()] = c
	}

	wrote := 0
	for _, group := range topLevelGroups {
		cmd, ok := byName[group]
		if !ok {
			fmt.Fprintf(os.Stderr, "warn: cobra has no top-level %q; skipping\n", group)
			continue
		}
		path := filepath.Join(docsDir, group+".md")
		if err := regenerate(path, cmd); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", path, err)
			os.Exit(1)
		}
		wrote++
	}
	fmt.Printf("regenerated %d per-group pages\n", wrote)
}

// regenerate rewrites the auto-generated header of a per-group docs
// page. Hand-written content below the AUTOGEN_END marker is
// preserved verbatim.
func regenerate(path string, cmd *cobra.Command) error {
	header := buildHeader(cmd)

	existing, err := os.ReadFile(path)
	tail := defaultTail()
	if err == nil {
		if idx := strings.Index(string(existing), autogenEnd); idx >= 0 {
			// Subsequent run: split on marker, preserve everything
			// after.
			afterMarker := string(existing[idx+len(autogenEnd):])
			tail = strings.TrimLeft(afterMarker, "\n")
		} else {
			// First run: no marker. The existing prose is being
			// converted. Preserve what's clearly hand-written
			// (Examples, conceptual blocks) by stripping the
			// auto-generatable parts (frontmatter, intro, the old
			// Subcommands table). The first heading that isn't
			// "## Subcommands" marks the start of preserved tail.
			tail = preserveTailFromUnmarked(string(existing))
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	out := header + "\n" + autogenEnd + "\n\n" + tail
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(out), 0o644)
}

func buildHeader(cmd *cobra.Command) string {
	var b strings.Builder

	// Frontmatter
	title := cmd.Name()
	desc := strings.TrimSpace(cmd.Short)
	if desc == "" {
		desc = title
	}
	fmt.Fprintf(&b, "---\ntitle: %s\ndescription: %s\n---\n\n", title, escapeYAML(desc))

	// Intro from cobra.Long (or fall back to Short)
	intro := strings.TrimSpace(cmd.Long)
	if intro == "" {
		intro = strings.TrimSpace(cmd.Short)
	}
	if intro != "" {
		fmt.Fprintf(&b, "%s\n\n", intro)
	}

	// Subcommands table - leaf commands get one row; nested groups
	// link via their own per-group pages elsewhere.
	subs := visibleSubcommands(cmd)
	if len(subs) > 0 {
		fmt.Fprintln(&b, "## Subcommands")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "| Command | Description |")
		fmt.Fprintln(&b, "| --- | --- |")
		for _, s := range subs {
			useLine := strings.TrimSpace(strings.TrimPrefix(s.Use, s.Name()))
			useLine = strings.TrimSpace(useLine)
			label := fmt.Sprintf("`%s %s`", cmd.Name(), s.Name())
			if useLine != "" && !strings.HasPrefix(useLine, "[") {
				label = fmt.Sprintf("`%s %s %s`", cmd.Name(), s.Name(), useLine)
			}
			short := strings.TrimSpace(s.Short)
			short = strings.ReplaceAll(short, "|", `\|`)
			fmt.Fprintf(&b, "| %s | %s |\n", label, short)
		}
		fmt.Fprintln(&b)
	}

	if len(cmd.Aliases) > 0 {
		fmt.Fprintf(&b, "Aliases: `%s`.\n\n", strings.Join(cmd.Aliases, "`, `"))
	}

	fmt.Fprintln(&b, "Full flag reference: see [the CLI reference](/reference/cli/).")

	return strings.TrimRight(b.String(), "\n")
}

func visibleSubcommands(cmd *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	for _, s := range cmd.Commands() {
		if s.Hidden || s.Name() == "help" {
			continue
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

func escapeYAML(s string) string {
	// Quote the description if it contains a colon (YAML key/value
	// separator hazard).
	if strings.Contains(s, ": ") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}

func defaultTail() string {
	return "## Examples\n\nAdd hand-written examples and conceptual notes below. Anything below the `AUTOGEN_END` marker is preserved across regenerations.\n"
}

// preserveTailFromUnmarked is the first-run conversion. The script
// finds the first `##` heading that isn't `## Subcommands`, and
// preserves everything from there onward. Frontmatter, intro
// paragraphs, and the existing Subcommands table are dropped (the
// regenerated header will provide all three fresh).
func preserveTailFromUnmarked(body string) string {
	lines := strings.Split(body, "\n")
	startIdx := -1
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "## ") {
			continue
		}
		if strings.EqualFold(trim, "## Subcommands") {
			continue
		}
		startIdx = i
		break
	}
	if startIdx < 0 {
		// No hand-written sections; use the default stub.
		return defaultTail()
	}
	return strings.TrimLeft(strings.Join(lines[startIdx:], "\n"), "\n")
}
