package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/robzolkos/rollbar-cli/internal/api"
)

func newOccurrenceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "occurrence <id>",
		Short: "Get single occurrence details",
		Long: `Get full details for a single occurrence (instance) by ID.

Examples:
  rollbar occurrence 123456789          # Get occurrence details
  rollbar occurrence 123456789 --ai     # AI-friendly output`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				return err
			}

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid occurrence ID: %w", err)
			}

			client := api.NewClient(cfg.AccessToken)
			instance, err := client.GetInstance(id)
			if err != nil {
				return err
			}

			formatter := getFormatter()
			return formatter.FormatInstance(os.Stdout, instance)
		},
	}

	return cmd
}
