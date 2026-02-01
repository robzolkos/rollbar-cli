package api

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLevelToString(t *testing.T) {
	tests := []struct {
		level    int
		expected string
	}{
		{10, "debug"},
		{20, "info"},
		{30, "warning"},
		{40, "error"},
		{50, "critical"},
		{0, "unknown"},
		{99, "unknown"},
	}

	for _, tt := range tests {
		result := LevelToString(tt.level)
		if result != tt.expected {
			t.Errorf("LevelToString(%d) = %s, want %s", tt.level, result, tt.expected)
		}
	}
}

func TestJSONLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{`"error"`, 40},
		{`"warning"`, 30},
		{`"info"`, 20},
		{`"debug"`, 10},
		{`"critical"`, 50},
		{`40`, 40},
		{`30`, 30},
	}

	for _, tt := range tests {
		var level JSONLevel
		if err := json.Unmarshal([]byte(tt.input), &level); err != nil {
			t.Errorf("failed to unmarshal %s: %v", tt.input, err)
			continue
		}
		if level.Int() != tt.expected {
			t.Errorf("JSONLevel.Int() for %s = %d, want %d", tt.input, level.Int(), tt.expected)
		}
	}
}

func TestJSONInt64(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`"123456789"`, 123456789},
		{`123456789`, 123456789},
		{`"0"`, 0},
		{`0`, 0},
	}

	for _, tt := range tests {
		var id JSONInt64
		if err := json.Unmarshal([]byte(tt.input), &id); err != nil {
			t.Errorf("failed to unmarshal %s: %v", tt.input, err)
			continue
		}
		if id.Int64() != tt.expected {
			t.Errorf("JSONInt64.Int64() for %s = %d, want %d", tt.input, id.Int64(), tt.expected)
		}
	}
}

func TestItemComputeFields(t *testing.T) {
	now := time.Now().Unix()
	item := Item{
		Level:                    JSONLevel(40),
		LastOccurrenceTimestamp:  now,
		FirstOccurrenceTimestamp: now - 3600, // 1 hour ago
	}

	item.ComputeFields()

	if item.LevelString != "error" {
		t.Errorf("expected LevelString 'error', got '%s'", item.LevelString)
	}

	if item.LastOccurrenceTime.IsZero() {
		t.Error("expected LastOccurrenceTime to be set")
	}

	if item.FirstOccurrenceTime.IsZero() {
		t.Error("expected FirstOccurrenceTime to be set")
	}
}

func TestInstanceComputeFields(t *testing.T) {
	now := time.Now().Unix()
	inst := Instance{
		Timestamp: now,
	}

	inst.ComputeFields()

	if inst.Time.IsZero() {
		t.Error("expected Time to be set")
	}

	// Should be within a second of now
	diff := time.Since(inst.Time)
	if diff > time.Second {
		t.Errorf("expected Time to be close to now, got %v ago", diff)
	}
}
