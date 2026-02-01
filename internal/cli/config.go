package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/robzolkos/rollbar-cli/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long: `Manage Rollbar CLI configuration.

Examples:
  rollbar config show                    # Show current config
  rollbar config set access_token <tok>  # Set access token
  rollbar config set project_id <id>     # Set project ID`,
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigSetCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg == nil {
				cfg = &config.Config{}
			}

			fmt.Fprintf(os.Stdout, "Config file: %s\n\n", getConfigSource())

			if cfg.AccessToken != "" {
				// Mask the token
				masked := cfg.AccessToken
				if len(masked) > 8 {
					masked = masked[:4] + "****" + masked[len(masked)-4:]
				}
				fmt.Fprintf(os.Stdout, "access_token: %s\n", masked)
			} else {
				fmt.Fprintln(os.Stdout, "access_token: (not set)")
			}

			if cfg.ProjectID != 0 {
				fmt.Fprintf(os.Stdout, "project_id: %d\n", cfg.ProjectID)
			}

			if cfg.DefaultEnvironment != "" {
				fmt.Fprintf(os.Stdout, "default_environment: %s\n", cfg.DefaultEnvironment)
			}

			fmt.Fprintf(os.Stdout, "output.format: %s\n", cfg.Output.Format)
			fmt.Fprintf(os.Stdout, "output.color: %s\n", cfg.Output.Color)

			return nil
		},
	}
}

func getConfigSource() string {
	if cfgFile != "" {
		return cfgFile
	}

	// Check for env var
	if os.Getenv("ROLLBAR_ACCESS_TOKEN") != "" {
		return "(environment variables)"
	}

	// Check for local config
	if _, err := os.Stat(".rollbar.yaml"); err == nil {
		return ".rollbar.yaml"
	}
	if _, err := os.Stat(".rollbar.json"); err == nil {
		return ".rollbar.json"
	}

	// Check for global config
	home, _ := os.UserHomeDir()
	globalPath := home + "/.config/rollbar/config.yaml"
	if _, err := os.Stat(globalPath); err == nil {
		return globalPath
	}

	return "(none)"
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a configuration value in the local .rollbar.yaml file.

Keys: access_token, project_id, default_environment, output.format, output.color`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			// Load or create config
			localCfg, _ := config.Load("")
			if localCfg == nil {
				localCfg = &config.Config{
					Output: config.OutputConfig{
						Format: "table",
						Color:  "auto",
					},
				}
			}

			switch key {
			case "access_token":
				localCfg.AccessToken = value
			case "project_id":
				var id int
				if _, err := fmt.Sscanf(value, "%d", &id); err != nil {
					return fmt.Errorf("invalid project_id: %s", value)
				}
				localCfg.ProjectID = id
			case "default_environment":
				localCfg.DefaultEnvironment = value
			case "output.format":
				localCfg.Output.Format = value
			case "output.color":
				localCfg.Output.Color = value
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}

			if err := localCfg.Save(config.ConfigPath()); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			if !quiet {
				fmt.Fprintf(os.Stdout, "Set %s in %s\n", key, config.ConfigPath())
			}
			return nil
		},
	}
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration in current directory",
		Long: `Create a .rollbar.yaml configuration file in the current directory.

This will prompt for your access token if not provided via flag.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := config.ConfigPath()

			// Check if already exists
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("config file already exists: %s", path)
			}

			newCfg := &config.Config{
				Output: config.OutputConfig{
					Format: "table",
					Color:  "auto",
				},
			}

			// Check for token from env
			if token := os.Getenv("ROLLBAR_ACCESS_TOKEN"); token != "" {
				newCfg.AccessToken = token
			}

			if err := newCfg.Save(path); err != nil {
				return fmt.Errorf("creating config file: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Created %s\n", path)
			if newCfg.AccessToken == "" {
				fmt.Fprintln(os.Stdout, "")
				fmt.Fprintln(os.Stdout, "Next steps:")
				fmt.Fprintln(os.Stdout, "  1. Get a read token from Rollbar: Project Settings > Access Tokens")
				fmt.Fprintln(os.Stdout, "  2. Run: rollbar config set access_token <your-token>")
				fmt.Fprintln(os.Stdout, "  3. Test: rollbar whoami")
			}

			return nil
		},
	}
}
