package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	// Set env var
	os.Setenv("ROLLBAR_ACCESS_TOKEN", "test-token")
	defer os.Unsetenv("ROLLBAR_ACCESS_TOKEN")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.AccessToken != "test-token" {
		t.Errorf("expected token 'test-token', got '%s'", cfg.AccessToken)
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create temp dir
	tmpDir := t.TempDir()

	// Create config file
	configPath := filepath.Join(tmpDir, ".rollbar.yaml")
	content := `access_token: file-token
project_id: 12345
default_environment: production
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.AccessToken != "file-token" {
		t.Errorf("expected token 'file-token', got '%s'", cfg.AccessToken)
	}
	if cfg.ProjectID != 12345 {
		t.Errorf("expected project_id 12345, got %d", cfg.ProjectID)
	}
	if cfg.DefaultEnvironment != "production" {
		t.Errorf("expected environment 'production', got '%s'", cfg.DefaultEnvironment)
	}
}

func TestLoadFromJSON(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, ".rollbar.json")
	content := `{"access_token": "json-token", "project_id": 999}`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.AccessToken != "json-token" {
		t.Errorf("expected token 'json-token', got '%s'", cfg.AccessToken)
	}
	if cfg.ProjectID != 999 {
		t.Errorf("expected project_id 999, got %d", cfg.ProjectID)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config file
	configPath := filepath.Join(tmpDir, ".rollbar.yaml")
	content := `access_token: file-token`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set env var (should override file)
	os.Setenv("ROLLBAR_ACCESS_TOKEN", "env-token")
	defer os.Unsetenv("ROLLBAR_ACCESS_TOKEN")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Env should override file
	if cfg.AccessToken != "env-token" {
		t.Errorf("expected token 'env-token' (from env), got '%s'", cfg.AccessToken)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     Config{AccessToken: "some-token"},
			wantErr: false,
		},
		{
			name:    "missing token",
			cfg:     Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".rollbar.yaml")

	cfg := &Config{
		AccessToken:        "my-token",
		ProjectID:          123,
		DefaultEnvironment: "staging",
		Output: OutputConfig{
			Format: "json",
			Color:  "never",
		},
	}

	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Read back
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Clear env to avoid interference
	os.Unsetenv("ROLLBAR_ACCESS_TOKEN")

	if loaded.AccessToken != "my-token" {
		t.Errorf("expected token 'my-token', got '%s'", loaded.AccessToken)
	}
	if loaded.ProjectID != 123 {
		t.Errorf("expected project_id 123, got %d", loaded.ProjectID)
	}
}
