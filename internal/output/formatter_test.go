package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/robzolkos/rollbar-cli/internal/api"
)

func sampleItems() []api.Item {
	now := time.Now()
	return []api.Item{
		{
			ID:                      api.JSONInt64(1),
			Counter:                 123,
			Title:                   "Test Error",
			Level:                   api.JSONLevel(40),
			LevelString:             "error",
			Status:                  "active",
			Environment:             "production",
			TotalOccurrences:        42,
			LastOccurrenceTime:      now,
			FirstOccurrenceTime:     now.Add(-24 * time.Hour),
			LastOccurrenceTimestamp: now.Unix(),
		},
	}
}

func sampleItem() *api.Item {
	items := sampleItems()
	return &items[0]
}

func sampleInstances() []api.Instance {
	now := time.Now()
	return []api.Instance{
		{
			ID:        999,
			Timestamp: now.Unix(),
			Time:      now,
			Data: api.InstanceData{
				Level:       "error",
				Environment: "production",
				Body: api.Body{
					Trace: &api.Trace{
						Exception: api.Exception{
							Class:   "TestError",
							Message: "Something went wrong",
						},
						Frames: []api.Frame{
							{Filename: "test.go", Lineno: 42, Method: "TestFunc"},
						},
					},
				},
				Request: &api.Request{
					URL:    "https://example.com/api/test",
					Method: "POST",
					Headers: map[string]string{
						"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
					},
					UserIP: "192.168.1.100",
				},
				Person: &api.Person{
					ID:    "user123",
					Email: "test@example.com",
				},
			},
		},
	}
}

func TestJSONFormatter(t *testing.T) {
	f := &JSONFormatter{}
	var buf bytes.Buffer

	t.Run("FormatItems", func(t *testing.T) {
		buf.Reset()
		err := f.FormatItems(&buf, sampleItems())
		if err != nil {
			t.Fatalf("FormatItems failed: %v", err)
		}

		var items []api.Item
		if err := json.Unmarshal(buf.Bytes(), &items); err != nil {
			t.Errorf("output is not valid JSON: %v", err)
		}
	})

	t.Run("FormatItem", func(t *testing.T) {
		buf.Reset()
		err := f.FormatItem(&buf, sampleItem())
		if err != nil {
			t.Fatalf("FormatItem failed: %v", err)
		}

		var item api.Item
		if err := json.Unmarshal(buf.Bytes(), &item); err != nil {
			t.Errorf("output is not valid JSON: %v", err)
		}
	})
}

func TestTableFormatter(t *testing.T) {
	f := &TableFormatter{Color: false}
	var buf bytes.Buffer

	t.Run("FormatItems", func(t *testing.T) {
		buf.Reset()
		err := f.FormatItems(&buf, sampleItems())
		if err != nil {
			t.Fatalf("FormatItems failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "TITLE") {
			t.Error("expected header 'TITLE' in table output")
		}
		if !strings.Contains(output, "123") {
			t.Error("expected item counter '123' in output")
		}
	})

	t.Run("FormatItems empty", func(t *testing.T) {
		buf.Reset()
		err := f.FormatItems(&buf, []api.Item{})
		if err != nil {
			t.Fatalf("FormatItems failed: %v", err)
		}

		if !strings.Contains(buf.String(), "No items found") {
			t.Error("expected 'No items found' message")
		}
	})
}

func TestCompactFormatter(t *testing.T) {
	f := &CompactFormatter{}
	var buf bytes.Buffer

	t.Run("FormatItems", func(t *testing.T) {
		buf.Reset()
		err := f.FormatItems(&buf, sampleItems())
		if err != nil {
			t.Fatalf("FormatItems failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "#123") {
			t.Error("expected item counter '#123' in compact output")
		}
		if !strings.Contains(output, "[error]") {
			t.Error("expected level '[error]' in compact output")
		}
	})
}

func TestMarkdownFormatter(t *testing.T) {
	f := &MarkdownFormatter{}
	var buf bytes.Buffer

	t.Run("FormatItems", func(t *testing.T) {
		buf.Reset()
		err := f.FormatItems(&buf, sampleItems())
		if err != nil {
			t.Fatalf("FormatItems failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "# Rollbar Items") {
			t.Error("expected markdown header '# Rollbar Items'")
		}
		if !strings.Contains(output, "|") {
			t.Error("expected markdown table with '|'")
		}
	})

	t.Run("FormatContext", func(t *testing.T) {
		buf.Reset()
		err := f.FormatContext(&buf, sampleItem(), sampleInstances())
		if err != nil {
			t.Fatalf("FormatContext failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "# Bug Report") {
			t.Error("expected '# Bug Report' header")
		}
		if !strings.Contains(output, "## Summary") {
			t.Error("expected '## Summary' section")
		}
		if !strings.Contains(output, "## Stack Trace") {
			t.Error("expected '## Stack Trace' section")
		}
	})
}

func TestCompactFormatterInstance(t *testing.T) {
	f := &CompactFormatter{}
	var buf bytes.Buffer

	t.Run("FormatInstance with browser and user", func(t *testing.T) {
		buf.Reset()
		inst := sampleInstances()[0]
		err := f.FormatInstance(&buf, &inst)
		if err != nil {
			t.Fatalf("FormatInstance failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Browser:") {
			t.Error("expected 'Browser:' in compact instance output")
		}
		if !strings.Contains(output, "Mozilla") {
			t.Error("expected User-Agent string in output")
		}
		if !strings.Contains(output, "test@example.com") {
			t.Error("expected user email in output")
		}
	})

	t.Run("FormatContext with browser and user", func(t *testing.T) {
		buf.Reset()
		err := f.FormatContext(&buf, sampleItem(), sampleInstances())
		if err != nil {
			t.Fatalf("FormatContext failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Browser:") {
			t.Error("expected 'Browser:' in compact context output")
		}
		if !strings.Contains(output, "## Person") {
			t.Error("expected '## Person' section in context output")
		}
		if !strings.Contains(output, "test@example.com") {
			t.Error("expected user email in context output")
		}
	})
}

func TestMarkdownFormatterInstance(t *testing.T) {
	f := &MarkdownFormatter{}
	var buf bytes.Buffer

	t.Run("FormatInstance with browser and user", func(t *testing.T) {
		buf.Reset()
		inst := sampleInstances()[0]
		err := f.FormatInstance(&buf, &inst)
		if err != nil {
			t.Fatalf("FormatInstance failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "**Browser:**") {
			t.Error("expected '**Browser:**' in markdown instance output")
		}
		if !strings.Contains(output, "Mozilla") {
			t.Error("expected User-Agent string in output")
		}
		if !strings.Contains(output, "## Person") {
			t.Error("expected '## Person' section")
		}
		if !strings.Contains(output, "test@example.com") {
			t.Error("expected user email in output")
		}
	})

	t.Run("FormatContext with browser in occurrences", func(t *testing.T) {
		buf.Reset()
		err := f.FormatContext(&buf, sampleItem(), sampleInstances())
		if err != nil {
			t.Fatalf("FormatContext failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "**Browser:**") {
			t.Error("expected '**Browser:**' in markdown context output")
		}
		if !strings.Contains(output, "**User:**") {
			t.Error("expected '**User:**' in markdown context output")
		}
	})
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		format   Format
		expected string
	}{
		{FormatJSON, "*output.JSONFormatter"},
		{FormatTable, "*output.TableFormatter"},
		{FormatCompact, "*output.CompactFormatter"},
		{FormatMarkdown, "*output.MarkdownFormatter"},
		{"unknown", "*output.TableFormatter"}, // defaults to table
	}

	for _, tt := range tests {
		f := New(tt.format, false)
		// Just verify it doesn't panic and returns a formatter
		if f == nil {
			t.Errorf("New(%s) returned nil", tt.format)
		}
	}
}
