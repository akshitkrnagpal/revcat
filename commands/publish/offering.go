package publish

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var (
	pubCurrent     bool
	pubNoCurrent   bool
	pubPaywallPath string
	pubConfirm     bool
	pubDryRun      bool
)

var offeringCmd = &cobra.Command{
	Use:   "offering <id>",
	Short: "Set an offering as current and/or push a paywall config",
	Long: `Composes the dashboard's "deploy" workflow:

  1. Verify the offering exists in the active project
  2. (optional) Validate and PUT a paywall config from --paywall <path>
  3. (optional) Promote the offering to current

By default both steps run when their inputs are provided. The plan is
printed before execution; pass --confirm to skip the prompt, or --dry-run
to print the plan without making any changes.

Examples:
    revcat publish offering default --current --confirm
    revcat publish offering pro --paywall ./paywalls/pro.json --current
    revcat publish offering pro --paywall ./paywalls/pro.json --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runPublishOffering,
}

func init() {
	offeringCmd.Flags().BoolVar(&pubCurrent, "current", false, "Set the offering as current")
	offeringCmd.Flags().BoolVar(&pubNoCurrent, "no-current", false, "Do NOT set the offering as current (overrides --current default when --paywall is set alone)")
	offeringCmd.Flags().StringVar(&pubPaywallPath, "paywall", "", "Path to a paywall config JSON file to PUT")
	offeringCmd.Flags().BoolVarP(&pubConfirm, "confirm", "y", false, "Skip the confirmation prompt")
	offeringCmd.Flags().BoolVar(&pubDryRun, "dry-run", false, "Print the plan without making changes")
}

func runPublishOffering(cmd *cobra.Command, args []string) error {
	id := args[0]

	client, _, err := cliutil.Client(cmd)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
	defer cancel()

	target, err := client.GetOffering(ctx, id, false)
	if err != nil {
		return fmt.Errorf("offering %q: %w", id, err)
	}

	// If neither --current nor --paywall is set explicitly, default to
	// --current=true. --no-current always wins.
	wantCurrent := pubCurrent
	if !wantCurrent && pubPaywallPath == "" {
		wantCurrent = true
	}
	if pubNoCurrent {
		wantCurrent = false
	}

	var paywallBody map[string]any
	if pubPaywallPath != "" {
		paywallBody, err = loadPaywall(pubPaywallPath)
		if err != nil {
			return err
		}
	}

	// Diff plan: what will actually change?
	steps := planSteps(target.IsCurrent, wantCurrent, paywallBody, func() (map[string]any, error) {
		return client.GetPaywall(ctx, id)
	})

	if len(steps) == 0 {
		output.Info("nothing to do: offering %q is already current and no paywall change requested", id)
		return nil
	}

	output.Info("plan for offering %q:", id)
	for _, s := range steps {
		output.Info("  %s %s", arrow(s.kind), s.desc)
	}

	if pubDryRun {
		output.Info("(dry run; no changes made)")
		return nil
	}

	if !pubConfirm {
		var ok bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "apply these changes?",
			Default: false,
		}, &ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("aborted")
		}
	}

	for _, s := range steps {
		switch s.kind {
		case stepPaywall:
			if err := client.PutPaywall(ctx, id, paywallBody); err != nil {
				return fmt.Errorf("put paywall: %w", err)
			}
			output.Success("paywall pushed (%d bytes)", s.size)
		case stepSetCurrent:
			if _, err := client.SetCurrentOffering(ctx, id); err != nil {
				return fmt.Errorf("set current: %w", err)
			}
			output.Success("offering %q is now current", id)
		}
	}
	return nil
}

type stepKind int

const (
	stepPaywall stepKind = iota
	stepSetCurrent
)

type step struct {
	kind stepKind
	desc string
	size int
}

func arrow(k stepKind) string {
	switch k {
	case stepPaywall:
		return "→"
	case stepSetCurrent:
		return "★"
	}
	return "·"
}

// planSteps figures out the minimal set of side effects needed. The
// existing paywall is fetched lazily (paywallFetcher closure) so we
// don't make a network call when --paywall isn't passed.
func planSteps(isCurrent, wantCurrent bool, newPaywall map[string]any, paywallFetcher func() (map[string]any, error)) []step {
	var steps []step

	if newPaywall != nil {
		newHash, newSize := hashJSON(newPaywall)
		desc := fmt.Sprintf("PUT paywall (%s, %d bytes)", newHash, newSize)
		emit := true
		if existing, err := paywallFetcher(); err == nil && existing != nil {
			oldHash, _ := hashJSON(existing)
			if oldHash == newHash {
				emit = false // identical, skip silently
			} else {
				desc = fmt.Sprintf("PUT paywall (%s -> %s, %d bytes)", oldHash, newHash, newSize)
			}
		}
		if emit {
			steps = append(steps, step{kind: stepPaywall, desc: desc, size: newSize})
		}
	}

	if wantCurrent && !isCurrent {
		steps = append(steps, step{kind: stepSetCurrent, desc: "promote to current"})
	}
	return steps
}

func loadPaywall(path string) (map[string]any, error) {
	b, err := cliutil.ReadCappedFile(path)
	if err != nil {
		return nil, err
	}
	var body map[string]any
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return body, nil
}

// hashJSON canonicalizes the value by re-marshalling with sorted keys so
// equivalent payloads compare equal regardless of key order.
func hashJSON(v map[string]any) (string, int) {
	b := canonicalJSON(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8]), len(b)
}

// canonicalJSON serializes a map in key-sorted order, recursively. Stable
// output makes hash comparisons meaningful.
func canonicalJSON(v any) []byte {
	switch x := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := []byte{'{'}
		for i, k := range keys {
			if i > 0 {
				out = append(out, ',')
			}
			kb, _ := json.Marshal(k)
			out = append(out, kb...)
			out = append(out, ':')
			out = append(out, canonicalJSON(x[k])...)
		}
		out = append(out, '}')
		return out
	case []any:
		out := []byte{'['}
		for i, item := range x {
			if i > 0 {
				out = append(out, ',')
			}
			out = append(out, canonicalJSON(item)...)
		}
		out = append(out, ']')
		return out
	default:
		b, _ := json.Marshal(x)
		return b
	}
}
