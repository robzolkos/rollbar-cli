package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/robzolkos/rollbar-cli/internal/api"
)

func newItemCmd() *cobra.Command {
	var (
		uuid        string
		occurrences int
		context     bool
	)

	cmd := &cobra.Command{
		Use:   "item <counter>",
		Short: "Get item details",
		Long: `Get details for a specific item by its project counter (e.g., #123).

Examples:
  rollbar item 123                    # Get item by counter
  rollbar item 123 --occurrences 5    # Include 5 recent occurrences
  rollbar item 123 --context          # Include full context (like 'context' command)
  rollbar item --uuid abc-def-123     # Get by UUID (internal ID)`,
		Args: func(cmd *cobra.Command, args []string) error {
			if uuid == "" && len(args) != 1 {
				return fmt.Errorf("requires item counter argument or --uuid flag")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				return err
			}

			client := api.NewClient(cfg.AccessToken)

			var item *api.Item
			var err error

			if uuid != "" {
				// Parse UUID as int64
				id, parseErr := strconv.ParseInt(uuid, 10, 64)
				if parseErr != nil {
					return fmt.Errorf("invalid UUID: %w", parseErr)
				}
				item, err = client.GetItem(id)
			} else {
				counter, parseErr := strconv.Atoi(args[0])
				if parseErr != nil {
					return fmt.Errorf("invalid counter: %w", parseErr)
				}
				item, err = client.GetItemByCounter(counter)
			}

			if err != nil {
				return err
			}

			formatter := getFormatter()

			// If context flag or occurrences requested, fetch instances
			if context || occurrences > 0 {
				instances, fetchErr := client.ListInstances(api.InstancesOptions{
					ItemID: item.ID.Int64(),
				})
				if fetchErr != nil {
					return fetchErr
				}

				if occurrences > 0 && len(instances) > occurrences {
					instances = instances[:occurrences]
				}

				return formatter.FormatContext(os.Stdout, item, instances)
			}

			return formatter.FormatItem(os.Stdout, item)
		},
	}

	cmd.Flags().StringVar(&uuid, "uuid", "", "get item by internal UUID/ID")
	cmd.Flags().IntVar(&occurrences, "occurrences", 0, "include N recent occurrences")
	cmd.Flags().BoolVar(&context, "context", false, "include full context (same as 'context' command)")

	return cmd
}
