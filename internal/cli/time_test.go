package cli

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		// Go duration format
		{"1h", time.Hour, false},
		{"24h", 24 * time.Hour, false},
		{"30m", 30 * time.Minute, false},

		// Human-friendly formats
		{"8 hours", 8 * time.Hour, false},
		{"8 hours ago", 8 * time.Hour, false},
		{"24 hours", 24 * time.Hour, false},
		{"7 days", 7 * 24 * time.Hour, false},
		{"7 days ago", 7 * 24 * time.Hour, false},
		{"2 weeks", 14 * 24 * time.Hour, false},
		{"1 month", 30 * 24 * time.Hour, false},

		// Short forms
		{"1 hr", time.Hour, false},
		{"1 min", time.Minute, false},
		{"1 d", 24 * time.Hour, false},
		{"1 w", 7 * 24 * time.Hour, false},

		// Invalid
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Result should be approximately now - expected duration
			expectedTime := now.Add(-tt.expected)
			diff := result.Sub(expectedTime)
			if diff < -time.Second || diff > time.Second {
				t.Errorf("result time %v differs from expected %v by %v", result, expectedTime, diff)
			}
		})
	}
}

func TestParseTimeArg(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2026-01-31T10:30:00", false},
		{"2026-01-31T10:30", false},
		{"2026-01-31", false},
		{"2026-01-31T10:30:00Z", false},

		// Invalid
		{"not a date", true},
		{"2026/01/31", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parseTimeArg(tt.input)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
