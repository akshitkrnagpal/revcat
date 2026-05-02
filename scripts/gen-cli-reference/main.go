// gen-cli-reference walks the cobra command tree and writes a single
// canonical CLI reference page to docs/src/content/docs/reference/cli.md.
//
// Why: hand-maintained docs (commands/<group>.md, skills/*) drift from
// the actual CLI surface. This generator produces a tier-2 derived doc
// that's always in sync with the code: rerun it after any command tree
// change and commit the resulting cli.md.
//
// Usage (from the repo root):
//
//	go run ./scripts/gen-cli-reference
//
// Or via the make target:
//
//	make docs-cli
//
// Output is one big markdown file with a section per command. Format
// is Astro Starlight compatible (frontmatter + headings).
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/akshitkrnagpal/revcat/commands"
)

const outputPath = "docs/src/content/docs/reference/cli.md"

const frontmatter = `---
title: CLI reference
description: Auto-generated reference for every revcat command. Regenerate with ` + "`go run ./scripts/gen-cli-reference`" + `.
---

This page is generated from the cobra command tree (` + "`go run ./scripts/gen-cli-reference`" + `). It's the canonical surface — when prose elsewhere disagrees, this page wins.

`

func main() {
	root := commands.RootCmd()

	var buf strings.Builder
	buf.WriteString(frontmatter)

	// Collect all commands depth-first, skipping the root (we render
	// its children at the top level).
	var all []*cobra.Command
	walk(root, func(c *cobra.Command) {
		if c == root {
			return
		}
		// Skip the auto-generated `help` and `completion` parents
		// (their subtrees are noise; we mention them in the intro).
		if c.Name() == "help" || c.Name() == "completion" {
			return
		}
		// Skip leaves under `completion` if the parent slipped through.
		if c.Parent() != nil && c.Parent().Name() == "completion" {
			return
		}
		all = append(all, c)
	})

	// Stable order: alphabetical by full command path.
	sort.SliceStable(all, func(i, j int) bool {
		return commandPath(all[i]) < commandPath(all[j])
	})

	for _, c := range all {
		writeCommand(&buf, c)
	}

	target := outputPath
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(target, []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}
	fmt.Println("wrote", target)
}

// walk visits every node in the cobra tree depth-first.
func walk(cmd *cobra.Command, fn func(*cobra.Command)) {
	fn(cmd)
	for _, child := range cmd.Commands() {
		walk(child, fn)
	}
}

// commandPath returns the dotted command path, e.g. "auth login".
func commandPath(c *cobra.Command) string {
	var parts []string
	for cur := c; cur != nil && cur.Parent() != nil; cur = cur.Parent() {
		parts = append([]string{cur.Name()}, parts...)
	}
	return strings.Join(parts, " ")
}

func writeCommand(out io.Writer, c *cobra.Command) {
	path := commandPath(c)
	fmt.Fprintf(out, "## `revcat %s`\n\n", path)

	if short := strings.TrimSpace(c.Short); short != "" {
		fmt.Fprintf(out, "%s\n\n", short)
	}

	if long := strings.TrimSpace(c.Long); long != "" && long != strings.TrimSpace(c.Short) {
		fmt.Fprintf(out, "```text\n%s\n```\n\n", long)
	}

	if use := strings.TrimSpace(c.UseLine()); use != "" {
		fmt.Fprintf(out, "**Usage**\n\n```sh\n%s\n```\n\n", use)
	}

	// Aliases
	if len(c.Aliases) > 0 {
		fmt.Fprintf(out, "**Aliases**: %s\n\n", strings.Join(quoteEach(c.Aliases), ", "))
	}

	// Local flags (excluding inherited globals - those are in their own
	// section at the bottom).
	if hasLocalFlags(c) {
		fmt.Fprintln(out, "**Flags**")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "| Flag | Type | Default | Description |")
		fmt.Fprintln(out, "| --- | --- | --- | --- |")
		c.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if f.Hidden {
				return
			}
			fmt.Fprintf(out, "| `%s` | %s | %s | %s |\n",
				flagSpelling(f),
				escape(f.Value.Type()),
				escape(defaultOrDash(f)),
				escape(strings.ReplaceAll(f.Usage, "\n", " ")),
			)
		})
		fmt.Fprintln(out)
	}

	// List subcommand names if this is a parent.
	if subs := visibleSubcommands(c); len(subs) > 0 {
		var names []string
		for _, s := range subs {
			names = append(names, "`"+s.Name()+"`")
		}
		fmt.Fprintf(out, "**Subcommands**: %s\n\n", strings.Join(names, ", "))
	}

	fmt.Fprintln(out, "---")
	fmt.Fprintln(out)
}

func hasLocalFlags(c *cobra.Command) bool {
	any := false
	c.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			any = true
		}
	})
	return any
}

func visibleSubcommands(c *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	for _, s := range c.Commands() {
		if s.Hidden || s.Name() == "help" {
			continue
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

func flagSpelling(f *pflag.Flag) string {
	if f.Shorthand != "" && f.Shorthand != " " {
		return fmt.Sprintf("-%s, --%s", f.Shorthand, f.Name)
	}
	return "--" + f.Name
}

func defaultOrDash(f *pflag.Flag) string {
	if f.DefValue == "" || f.DefValue == "false" || f.DefValue == "[]" {
		return "-"
	}
	return f.DefValue
}

func escape(s string) string {
	// Pipe characters break markdown tables.
	return strings.ReplaceAll(s, "|", `\|`)
}

func quoteEach(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = "`" + s + "`"
	}
	return out
}
