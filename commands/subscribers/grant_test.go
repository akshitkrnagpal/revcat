package subscribers

import (
	"testing"
	"time"
)

func TestDurationToExpiresAt(t *testing.T) {
	now := time.Now().UnixMilli()
	day := int64(24 * time.Hour / time.Millisecond)

	cases := []struct {
		in        string
		minDays   int64
		maxDays   int64
		wantLabel string
		wantErr   bool
		lifetime  bool
	}{
		{in: "7d", minDays: 7, maxDays: 7, wantLabel: "7 day(s)"},
		{in: "30d", minDays: 30, maxDays: 30, wantLabel: "30 day(s)"},
		// AddDate respects calendar months, so 1m is 28-31 days depending on the start.
		{in: "1m", minDays: 28, maxDays: 31, wantLabel: "1 month(s)"},
		{in: "1y", minDays: 365, maxDays: 366, wantLabel: "1 year(s)"},
		{in: "forever", lifetime: true, wantLabel: "lifetime"},
		{in: "LIFETIME", lifetime: true, wantLabel: "lifetime"},
		{in: "", wantErr: true},
		{in: "x", wantErr: true},
		{in: "0d", wantErr: true},
		{in: "-1d", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			expires, label, err := durationToExpiresAt(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got %d / %q", expires, label)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if label != tc.wantLabel {
				t.Fatalf("label: want %q, got %q", tc.wantLabel, label)
			}
			delta := expires - now
			if tc.lifetime {
				if delta < int64(99*365)*day {
					t.Fatalf("lifetime expiry not far enough out (delta=%d ms)", delta)
				}
				return
			}
			minMS := tc.minDays*day - 5_000
			maxMS := tc.maxDays*day + 5_000
			if delta < minMS || delta > maxMS {
				t.Fatalf("delta=%d ms outside [%d, %d]", delta, minMS, maxMS)
			}
		})
	}
}
