package cliutil

import "time"

// FormatTime renders a unix timestamp (seconds OR ms; auto-detected) as
// YYYY-MM-DD. Used everywhere we surface a created_at column.
func FormatTime(unix int64) string {
	if unix == 0 {
		return "-"
	}
	t := time.Unix(unix, 0).UTC()
	if unix > 9999999999 {
		t = time.UnixMilli(unix).UTC()
	}
	return t.Format("2006-01-02")
}

// Dash returns "-" for empty strings, the value otherwise. Convenience
// for table cells.
func Dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
