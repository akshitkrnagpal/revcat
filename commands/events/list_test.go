package events

import (
	"testing"
	"time"
)

func TestParseSince(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got, err := parseSince("")
		if err != nil {
			t.Fatal(err)
		}
		if got != 0 {
			t.Fatalf("want 0, got %d", got)
		}
	})

	t.Run("rfc3339", func(t *testing.T) {
		got, err := parseSince("2026-04-25T10:00:00Z")
		if err != nil {
			t.Fatal(err)
		}
		want := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC).UnixMilli()
		if got != want {
			t.Fatalf("want %d, got %d", want, got)
		}
	})

	t.Run("duration with d", func(t *testing.T) {
		now := time.Now().UnixMilli()
		got, err := parseSince("7d")
		if err != nil {
			t.Fatal(err)
		}
		// Allow a 5s window for clock drift between the two Now() calls.
		delta := now - got
		want := int64(7 * 24 * 3600 * 1000)
		if delta < want-5000 || delta > want+5000 {
			t.Fatalf("want ~%d ago, got delta %d", want, delta)
		}
	})

	t.Run("duration with h", func(t *testing.T) {
		now := time.Now().UnixMilli()
		got, err := parseSince("1h")
		if err != nil {
			t.Fatal(err)
		}
		delta := now - got
		want := int64(3600 * 1000)
		if delta < want-5000 || delta > want+5000 {
			t.Fatalf("want ~%d ago, got delta %d", want, delta)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if _, err := parseSince("not-a-time"); err == nil {
			t.Fatal("want error")
		}
	})
}
