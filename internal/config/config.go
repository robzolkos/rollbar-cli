package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration
type Config struct {
	AccessToken        string       `yaml:"access_token" json:"access_token"`
	ProjectID          int          `yaml:"project_id" json:"project_id"`
	DefaultEnvironment string       `yaml:"default_environment" json:"default_environment"`
	Output             OutputConfig `yaml:"output" json:"output"`
}

// OutputConfig configures output formatting
type OutputConfig struct {
	Format string `yaml:"format" json:"format"` // compact | json | markdown | table
	Color  string `yaml:"color" json:"color"`   // auto | always | never
}

// GlobalConfig represents the global ~/.config/rollbar/config.yaml format
type GlobalConfig struct {
	Profiles       map[string]Profile `yaml:"profiles" json:"profiles"`
	DefaultProfile string             `yaml:"default_profile" json:"default_profile"`
	Projects       map[string]struct {
		Profile   string `yaml:"profile" json:"profile"`
		ProjectID int    `yaml:"project_id" json:"project_id"`
	} `yaml:"projects" json:"projects"`
}

// Profile represents a named configuration profile
type Profile struct {
	AccessToken  string `yaml:"access_token" json:"access_token"`
	AccountToken string `yaml:"account_token" json:"account_token"`
}

// Load loads configuration with priority:
// 1. --config flag (explicit path)
// 2. .rollbar.yaml / .rollbar.json in current directory
// 3. Walk up directory tree looking for .rollbar.yaml
// 4. ~/.config/rollbar/config.yaml (global default)
// 5. Environment variables
func Load(configPath string) (*Config, error) {
	cfg := &Config{
		Output: OutputConfig{
			Format: "table",
			Color:  "auto",
		},
	}

	// Try explicit config path first
	if configPath != "" {
		if err := loadConfigFile(configPath, cfg); err != nil {
			return nil, fmt.Errorf("loading config from %s: %w", configPath, err)
		}
		applyEnvOverrides(cfg)
		return cfg, nil
	}

	// Try to find config file
	if path := discoverConfigFile(); path != "" {
		if err := loadConfigFile(path, cfg); err != nil {
			return nil, fmt.Errorf("loading config from %s: %w", path, err)
		}
	}

	// Try global config
	if cfg.AccessToken == "" {
		globalPath := globalConfigPath()
		if _, err := os.Stat(globalPath); err == nil {
			// Non-fatal error: just continue with env vars if loading fails
			_ = loadGlobalConfig(globalPath, cfg)
		}
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	return cfg, nil
}

// discoverConfigFile walks up directory tree looking for config files
func discoverConfigFile() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		// Check for .rollbar.yaml
		yamlPath := filepath.Join(dir, ".rollbar.yaml")
		if _, err := os.Stat(yamlPath); err == nil {
			return yamlPath
		}

		// Check for .rollbar.json
		jsonPath := filepath.Join(dir, ".rollbar.json")
		if _, err := os.Stat(jsonPath); err == nil {
			return jsonPath
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}

	return ""
}

func globalConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "rollbar", "config.yaml")
}

func loadConfigFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, cfg)
	case ".json":
		return json.Unmarshal(data, cfg)
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return json.Unmarshal(data, cfg)
		}
		return nil
	}
}

func loadGlobalConfig(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var global GlobalConfig
	if err := yaml.Unmarshal(data, &global); err != nil {
		return err
	}

	// Use default profile
	profileName := global.DefaultProfile
	if profileName == "" && len(global.Profiles) > 0 {
		for name := range global.Profiles {
			profileName = name
			break
		}
	}

	if profile, ok := global.Profiles[profileName]; ok {
		if cfg.AccessToken == "" {
			cfg.AccessToken = profile.AccessToken
		}
	}

	// Check for project-specific config based on cwd
	cwd, err := os.Getwd()
	if err == nil {
		for projectPath, projectCfg := range global.Projects {
			if isSubPath(projectPath, cwd) {
				if profile, ok := global.Profiles[projectCfg.Profile]; ok {
					cfg.AccessToken = profile.AccessToken
				}
				if projectCfg.ProjectID != 0 {
					cfg.ProjectID = projectCfg.ProjectID
				}
				break
			}
		}
	}

	return nil
}

func isSubPath(base, path string) bool {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return false
	}
	return rel != ".." && !filepath.IsAbs(rel) && rel[0] != '.'
}

func applyEnvOverrides(cfg *Config) {
	if token := os.Getenv("ROLLBAR_ACCESS_TOKEN"); token != "" {
		cfg.AccessToken = token
	}
	if env := os.Getenv("ROLLBAR_ENVIRONMENT"); env != "" {
		cfg.DefaultEnvironment = env
	}
}

// Validate checks if the configuration is valid for API calls
func (c *Config) Validate() error {
	if c.AccessToken == "" {
		return fmt.Errorf("access token not configured. Set ROLLBAR_ACCESS_TOKEN or create .rollbar.yaml")
	}
	return nil
}

// ConfigPath returns the path where a project config would be saved
func ConfigPath() string {
	return ".rollbar.yaml"
}

// Save saves the config to a file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
