package subscribers

import "testing"

func TestNormalizeDuration(t *testing.T) {
	cases := []struct {
		in        string
		wantRC    string
		wantLabel string
		wantErr   bool
	}{
		{"7d", "P7D", "7 day(s)", false},
		{"30d", "P30D", "30 day(s)", false},
		{"1m", "P1M", "1 month(s)", false},
		{"3m", "P3M", "3 month(s)", false},
		{"1y", "P1Y", "1 year(s)", false},
		{"forever", "lifetime", "lifetime", false},
		{"LIFETIME", "lifetime", "lifetime", false},
		{"monthly", "monthly", "monthly", false},
		{"yearly", "yearly", "yearly", false},
		{"", "", "", true},
		{"x", "", "", true},
		{"7x", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			rc, label, err := normalizeDuration(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got rc=%q label=%q", rc, label)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rc != tc.wantRC || label != tc.wantLabel {
				t.Fatalf("got (%q, %q), want (%q, %q)", rc, label, tc.wantRC, tc.wantLabel)
			}
		})
	}
}
