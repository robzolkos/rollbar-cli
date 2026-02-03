//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

var (
	token       string
	itemCounter int
)

func TestMain(m *testing.M) {
	token = os.Getenv("ROLLBAR_E2E_TOKEN")
	if token == "" {
		panic("ROLLBAR_E2E_TOKEN environment variable not set")
	}

	if counter := os.Getenv("ROLLBAR_E2E_ITEM_COUNTER"); counter != "" {
		var err error
		itemCounter, err = strconv.Atoi(counter)
		if err != nil {
			panic("invalid ROLLBAR_E2E_ITEM_COUNTER: " + err.Error())
		}
	}

	os.Exit(m.Run())
}

func runRollbar(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	// Build the command
	cmd := exec.Command("go", append([]string{"run", "../cmd/rollbar"}, args...)...)
	cmd.Env = append(os.Environ(), "ROLLBAR_ACCESS_TOKEN="+token)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestE2E_Whoami(t *testing.T) {
	stdout, stderr, err := runRollbar(t, "whoami")
	if err != nil {
		t.Fatalf("whoami failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "Project:") {
		t.Errorf("expected 'Project:' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "OK") && !strings.Contains(stdout, "Authentication") {
		t.Errorf("expected authentication confirmation in output, got: %s", stdout)
	}
}

func TestE2E_ItemsList(t *testing.T) {
	stdout, stderr, err := runRollbar(t, "items", "--output", "json")
	if err != nil {
		t.Fatalf("items failed: %v\nstderr: %s", err, stderr)
	}

	// Should be valid JSON array
	var items []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &items); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %s", err, stdout)
	}

	t.Logf("Found %d items", len(items))
}

func TestE2E_ItemsWithFilters(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"status active", []string{"items", "--status", "active", "--output", "json"}},
		{"level error", []string{"items", "--level", "error", "--output", "json"}},
		{"level critical,error", []string{"items", "--level", "critical,error", "--output", "json"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runRollbar(t, tt.args...)
			if err != nil {
				t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
			}

			var items []map[string]interface{}
			if err := json.Unmarshal([]byte(stdout), &items); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
			t.Logf("Found %d items with filter %s", len(items), tt.name)
		})
	}
}

func TestE2E_ItemsTimeFilters(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"since 24h", []string{"items", "--since", "24h", "--output", "json"}},
		{"since 7 days", []string{"items", "--since", "7 days", "--output", "json"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runRollbar(t, tt.args...)
			if err != nil {
				t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
			}

			var items []map[string]interface{}
			if err := json.Unmarshal([]byte(stdout), &items); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
			t.Logf("Found %d items with %s", len(items), tt.name)
		})
	}
}

func TestE2E_ItemsSearch(t *testing.T) {
	stdout, stderr, err := runRollbar(t, "items", "--query", "Error", "--output", "json")
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}

	var items []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &items); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	t.Logf("Found %d items matching 'Error'", len(items))
}

func TestE2E_ItemsSorting(t *testing.T) {
	tests := []struct {
		name string
		sort string
	}{
		{"recent", "recent"},
		{"occurrences", "occurrences"},
		{"level", "level"},
		{"first-seen", "first-seen"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runRollbar(t, "items", "--sort", tt.sort, "--output", "json")
			if err != nil {
				t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
			}

			var items []map[string]interface{}
			if err := json.Unmarshal([]byte(stdout), &items); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
			t.Logf("Sorted by %s: %d items", tt.sort, len(items))
		})
	}
}

func TestE2E_ItemDetail(t *testing.T) {
	if itemCounter == 0 {
		t.Skip("ROLLBAR_E2E_ITEM_COUNTER not set")
	}

	stdout, stderr, err := runRollbar(t, "item", strconv.Itoa(itemCounter), "--output", "json")
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}

	var item map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &item); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if item["counter"] == nil {
		t.Error("expected 'counter' field in item")
	}
	if item["title"] == nil {
		t.Error("expected 'title' field in item")
	}

	t.Logf("Item #%d: %s", itemCounter, item["title"])
}

func TestE2E_Occurrences(t *testing.T) {
	if itemCounter == 0 {
		t.Skip("ROLLBAR_E2E_ITEM_COUNTER not set")
	}

	stdout, stderr, err := runRollbar(t, "occurrences", "--item", strconv.Itoa(itemCounter), "--output", "json")
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}

	var instances []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &instances); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	t.Logf("Found %d occurrences for item #%d", len(instances), itemCounter)
}

func TestE2E_Context(t *testing.T) {
	if itemCounter == 0 {
		t.Skip("ROLLBAR_E2E_ITEM_COUNTER not set")
	}

	stdout, stderr, err := runRollbar(t, "context", strconv.Itoa(itemCounter))
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}

	// Should contain markdown headers
	if !strings.Contains(stdout, "# Bug Report") && !strings.Contains(stdout, "# Error") {
		t.Error("expected markdown header in context output")
	}
	if !strings.Contains(stdout, "## Summary") {
		t.Error("expected '## Summary' in context output")
	}

	t.Logf("Context output length: %d bytes", len(stdout))
}

func TestE2E_OutputFormats(t *testing.T) {
	tests := []struct {
		format   string
		validate func(string) bool
	}{
		{"json", func(s string) bool {
			var v interface{}
			return json.Unmarshal([]byte(s), &v) == nil
		}},
		{"table", func(s string) bool {
			return strings.Contains(s, "TITLE") || strings.Contains(s, "No items")
		}},
		{"compact", func(s string) bool {
			return strings.Contains(s, "#") || len(s) == 0
		}},
		{"markdown", func(s string) bool {
			return strings.Contains(s, "#") || strings.Contains(s, "|")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			stdout, stderr, err := runRollbar(t, "items", "--output", tt.format, "--limit", "5")
			if err != nil {
				t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
			}

			if !tt.validate(stdout) {
				t.Errorf("output validation failed for format %s:\n%s", tt.format, stdout)
			}
		})
	}
}

func TestE2E_InvalidToken(t *testing.T) {
	cmd := exec.Command("go", "run", "../cmd/rollbar", "whoami")
	cmd.Env = append(os.Environ(), "ROLLBAR_ACCESS_TOKEN=invalid-token-12345")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Error("expected error with invalid token")
	}
}

func TestE2E_AIMode(t *testing.T) {
	stdout, stderr, err := runRollbar(t, "items", "--ai", "--limit", "3")
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}

	// AI mode should use compact format with no ANSI codes
	if strings.Contains(stdout, "\033[") {
		t.Error("AI mode should not contain ANSI color codes")
	}

	t.Logf("AI mode output:\n%s", stdout)
}

func TestE2E_Resolve(t *testing.T) {
	if itemCounter == 0 {
		t.Skip("ROLLBAR_E2E_ITEM_COUNTER not set")
	}

	// First, get the item's current status
	stdout, stderr, err := runRollbar(t, "item", strconv.Itoa(itemCounter), "--output", "json")
	if err != nil {
		t.Fatalf("failed to get item: %v\nstderr: %s", err, stderr)
	}

	var item map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &item); err != nil {
		t.Fatalf("failed to parse item JSON: %v", err)
	}

	originalStatus := item["status"].(string)
	t.Logf("Item #%d original status: %s", itemCounter, originalStatus)

	// Resolve the item
	_, stderr, err = runRollbar(t, "resolve", strconv.Itoa(itemCounter))
	if err != nil {
		// Check if it's a permission error (read-only token)
		if strings.Contains(stderr, "scope") || strings.Contains(stderr, "403") || strings.Contains(stderr, "401") {
			t.Skip("Token doesn't have write permissions - skipping resolve test")
		}
		t.Fatalf("resolve failed: %v\nstderr: %s", err, stderr)
	}

	// Verify the item is now resolved
	stdout, stderr, err = runRollbar(t, "item", strconv.Itoa(itemCounter), "--output", "json")
	if err != nil {
		t.Fatalf("failed to get item after resolve: %v\nstderr: %s", err, stderr)
	}

	if err := json.Unmarshal([]byte(stdout), &item); err != nil {
		t.Fatalf("failed to parse item JSON: %v", err)
	}

	newStatus := item["status"].(string)
	if newStatus != "resolved" {
		t.Errorf("expected status 'resolved', got '%s'", newStatus)
	}

	t.Logf("Item #%d successfully resolved (was: %s, now: %s)", itemCounter, originalStatus, newStatus)
}

func TestE2E_ResolveMultiple(t *testing.T) {
	// Get two active items to resolve
	stdout, stderr, err := runRollbar(t, "items", "--status", "active", "--limit", "2", "--output", "json")
	if err != nil {
		t.Fatalf("failed to list items: %v\nstderr: %s", err, stderr)
	}

	var items []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &items); err != nil {
		t.Fatalf("failed to parse items JSON: %v", err)
	}

	if len(items) < 2 {
		t.Skip("Need at least 2 active items to test batch resolve")
	}

	counter1 := int(items[0]["counter"].(float64))
	counter2 := int(items[1]["counter"].(float64))

	t.Logf("Attempting to resolve items #%d and #%d", counter1, counter2)

	// Resolve both items
	_, stderr, err = runRollbar(t, "resolve", strconv.Itoa(counter1), strconv.Itoa(counter2))
	if err != nil {
		if strings.Contains(stderr, "scope") || strings.Contains(stderr, "403") || strings.Contains(stderr, "401") {
			t.Skip("Token doesn't have write permissions - skipping batch resolve test")
		}
		t.Fatalf("batch resolve failed: %v\nstderr: %s", err, stderr)
	}

	// Verify both items are resolved
	for _, counter := range []int{counter1, counter2} {
		stdout, stderr, err = runRollbar(t, "item", strconv.Itoa(counter), "--output", "json")
		if err != nil {
			t.Fatalf("failed to get item #%d: %v\nstderr: %s", counter, err, stderr)
		}

		var item map[string]interface{}
		if err := json.Unmarshal([]byte(stdout), &item); err != nil {
			t.Fatalf("failed to parse item JSON: %v", err)
		}

		if item["status"].(string) != "resolved" {
			t.Errorf("item #%d: expected status 'resolved', got '%s'", counter, item["status"])
		}
	}

	t.Logf("Successfully resolved items #%d and #%d", counter1, counter2)
}

func TestE2E_ResolveInvalidCounter(t *testing.T) {
	_, stderr, err := runRollbar(t, "resolve", "not-a-number")
	if err == nil {
		t.Error("expected error for invalid counter")
	}
	if !strings.Contains(stderr, "invalid counter") {
		t.Errorf("expected 'invalid counter' in error, got: %s", stderr)
	}
}

func TestE2E_ResolveNoArgs(t *testing.T) {
	_, stderr, err := runRollbar(t, "resolve")
	if err == nil {
		t.Error("expected error when no arguments provided")
	}
	if !strings.Contains(stderr, "requires") {
		t.Errorf("expected 'requires' in error message, got: %s", stderr)
	}
}
