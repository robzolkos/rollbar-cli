package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateItemStatus(t *testing.T) {
	tests := []struct {
		name           string
		itemID         int64
		status         string
		serverResponse interface{}
		serverStatus   int
		wantErr        bool
		errContains    string
	}{
		{
			name:   "resolve item successfully",
			itemID: 123,
			status: "resolved",
			serverResponse: map[string]interface{}{
				"err": 0,
				"result": map[string]interface{}{
					"id":      123,
					"counter": 42,
					"title":   "Test Error",
					"status":  "resolved",
					"level":   40,
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:   "mute item successfully",
			itemID: 456,
			status: "muted",
			serverResponse: map[string]interface{}{
				"err": 0,
				"result": map[string]interface{}{
					"id":      456,
					"counter": 99,
					"title":   "Another Error",
					"status":  "muted",
					"level":   30,
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:   "item not found",
			itemID: 999,
			status: "resolved",
			serverResponse: map[string]interface{}{
				"err":     1,
				"message": "Item not found",
			},
			serverStatus: http.StatusNotFound,
			wantErr:      true,
			errContains:  "Item not found",
		},
		{
			name:   "unauthorized - read-only token",
			itemID: 123,
			status: "resolved",
			serverResponse: map[string]interface{}{
				"err":     1,
				"message": "access token doesn't have the required scope",
			},
			serverStatus: http.StatusForbidden,
			wantErr:      true,
			errContains:  "scope",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "PATCH" {
					t.Errorf("expected PATCH request, got %s", r.Method)
				}

				// Verify content type
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				// Verify auth header
				if r.Header.Get("X-Rollbar-Access-Token") == "" {
					t.Error("expected X-Rollbar-Access-Token header")
				}

				// Decode request body
				var payload map[string]string
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Errorf("failed to decode request body: %v", err)
				}
				if payload["status"] != tt.status {
					t.Errorf("expected status %s, got %s", tt.status, payload["status"])
				}

				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client := NewClient("test-token")
			client.baseURL = server.URL

			item, err := client.UpdateItemStatus(tt.itemID, tt.status)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errContains != "" {
					if apiErr, ok := err.(*APIError); ok {
						if apiErr.Message == "" || !contains(apiErr.Message, tt.errContains) {
							t.Errorf("expected error containing %q, got %q", tt.errContains, apiErr.Message)
						}
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if item == nil {
				t.Fatal("expected item, got nil")
			}
		})
	}
}

func TestDoRequestWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("X-Rollbar-Access-Token") != "test-token" {
			t.Errorf("expected token 'test-token', got %s", r.Header.Get("X-Rollbar-Access-Token"))
		}

		// Echo back the body
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"err":    0,
			"result": body,
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	payload := map[string]string{"key": "value"}
	resp, err := client.doRequestWithBody("POST", "/test", nil, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["err"].(float64) != 0 {
		t.Error("expected err=0 in response")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
