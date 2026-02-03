package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/robzolkos/rollbar-cli/internal/config"
	"github.com/robzolkos/rollbar-cli/internal/output"
	"github.com/robzolkos/rollbar-cli/internal/version"
)

var (
	cfgFile      string
	outputFormat string
	aiMode       bool
	noColor      bool
	quiet        bool

	cfg *config.Config
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "rollbar",
	Short: "CLI for Rollbar error tracking",
	Long: `A command-line interface for Rollbar focused on reading, listing, and
managing items and occurrences. Optimized for both AI coding agents and human users.

Use 'rollbar items' to list errors, 'rollbar item <counter>' to get details,
'rollbar context <counter>' to generate AI-friendly bug context, and
'rollbar resolve <counter>' to mark items as resolved.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for version and completion commands
		if cmd.Name() == "version" || cmd.Name() == "completion" || cmd.Parent().Name() == "completion" {
			return nil
		}

		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return err
		}

		// Apply --ai flag shortcuts
		if aiMode {
			outputFormat = "compact"
			noColor = true
		}

		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .rollbar.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format: table, json, compact, markdown")
	rootCmd.PersistentFlags().BoolVar(&aiMode, "ai", false, "AI mode: shorthand for --output=compact --no-color")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")

	// Add subcommands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newItemsCmd())
	rootCmd.AddCommand(newItemCmd())
	rootCmd.AddCommand(newOccurrencesCmd())
	rootCmd.AddCommand(newOccurrenceCmd())
	rootCmd.AddCommand(newContextCmd())
	rootCmd.AddCommand(newWhoamiCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newResolveCmd())
}

// getFormatter returns the appropriate formatter based on flags
func getFormatter() output.Formatter {
	format := output.Format(outputFormat)
	useColor := !noColor && isTerminal()
	return output.New(format, useColor)
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// newVersionCmd creates the version command
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version.Full())
		},
	}
}
