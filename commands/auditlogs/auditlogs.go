// Package auditlogs holds `revcat audit-logs ...`.
package auditlogs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var Cmd = &cobra.Command{
	Use:     "audit-logs",
	Aliases: []string{"audit"},
	Short:   "Inspect the project's audit log",
}

func init() {
	Cmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List audit log entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		entries, err := client.ListAuditLogs(ctx)
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(entries)
		}
		rows := make([][]any, 0, len(entries))
		for _, e := range entries {
			actor := actorString(e)
			target := targetString(e)
			rows = append(rows, []any{cliutil.FormatTime(e.OccurredAt), e.ActionType, actor, target})
		}
		return output.Table([]string{"when", "action", "actor", "target"}, rows)
	},
}

// actorString tries to render a human-friendly actor (e.g., "Akshit
// Kr Nagpal" from additional_data.actor.name) and falls back to the
// raw type+identifier.
func actorString(e api.AuditLogEntry) string {
	if a, ok := e.AdditionalData["actor"].(map[string]any); ok {
		if name, ok := a["name"].(string); ok && name != "" {
			return name
		}
		if email, ok := a["email"].(string); ok && email != "" {
			return email
		}
	}
	if e.ActorType != "" {
		return e.ActorType + " " + cliutil.Dash(e.ActorIdentifier)
	}
	return cliutil.Dash(e.ActorIdentifier)
}

func targetString(e api.AuditLogEntry) string {
	if t, ok := e.AdditionalData["target"].(map[string]any); ok {
		if label, ok := t["label"].(string); ok && label != "" {
			if scope, ok := t["scope"].(string); ok && scope != "" {
				return label + " (" + scope + ")"
			}
			return label
		}
	}
	if e.TargetType != "" {
		return e.TargetType + " " + cliutil.Dash(e.TargetIdentifier)
	}
	return cliutil.Dash(e.TargetIdentifier)
}

// silenced unused-import shim if json + fmt aren't needed
var _ = json.Marshal
var _ = fmt.Sprintf
