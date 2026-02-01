package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/robzolkos/rollbar-cli/internal/api"
	"github.com/robzolkos/rollbar-cli/internal/output"
)

func newContextCmd() *cobra.Command {
	var (
		occurrences int
		outFile     string
	)

	cmd := &cobra.Command{
		Use:   "context <counter>",
		Short: "Generate AI context file for a bug",
		Long: `Generate a comprehensive context file with all information needed to fix a bug.
This is the primary command for AI agents.

Examples:
  rollbar context 123                          # Output to stdout
  rollbar context 123 --out bug-context.md     # Write to file
  rollbar context 123 --occurrences 5          # Include 5 recent occurrences
  rollbar context 123 | pbcopy                 # Copy to clipboard (macOS)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				return err
			}

			counter, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid counter: %w", err)
			}

			client := api.NewClient(cfg.AccessToken)

			// Get item details
			item, err := client.GetItemByCounter(counter)
			if err != nil {
				return err
			}

			// Get recent occurrences
			instances, err := client.ListInstances(api.InstancesOptions{
				ItemID: item.ID.Int64(),
			})
			if err != nil {
				return err
			}

			// Limit occurrences
			if occurrences > 0 && len(instances) > occurrences {
				instances = instances[:occurrences]
			} else if occurrences == 0 && len(instances) > 3 {
				// Default to 3 occurrences
				instances = instances[:3]
			}

			// Use markdown formatter for context (or JSON if specified)
			var formatter output.Formatter
			switch output.Format(outputFormat) {
			case output.FormatJSON:
				formatter = &output.JSONFormatter{}
			case output.FormatCompact:
				formatter = &output.CompactFormatter{}
			default:
				formatter = &output.MarkdownFormatter{}
			}

			// Write to file or stdout
			writer := os.Stdout
			if outFile != "" {
				f, createErr := os.Create(outFile)
				if createErr != nil {
					return fmt.Errorf("creating output file: %w", createErr)
				}
				defer f.Close()
				writer = f

				if !quiet {
					fmt.Fprintf(os.Stderr, "Writing context to %s\n", outFile)
				}
			}

			return formatter.FormatContext(writer, item, instances)
		},
	}

	cmd.Flags().IntVar(&occurrences, "occurrences", 3, "number of recent occurrences to include")
	cmd.Flags().StringVar(&outFile, "out", "", "output file path")

	return cmd
}
