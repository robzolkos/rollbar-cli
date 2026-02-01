package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/robzolkos/rollbar-cli/internal/api"
)

func newOccurrencesCmd() *cobra.Command {
	var (
		itemCounter int
		all         bool
		since       string
		limit       int
		page        int
	)

	cmd := &cobra.Command{
		Use:   "occurrences",
		Short: "List occurrences (instances)",
		Long: `List occurrences (individual error instances) for an item or all items.

Examples:
  rollbar occurrences --item 123         # List occurrences for item #123
  rollbar occurrences --all              # List all project occurrences
  rollbar occurrences --item 123 --limit 10   # Limit to 10 occurrences`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				return err
			}

			if !all && itemCounter == 0 {
				return fmt.Errorf("specify --item <counter> or --all")
			}

			client := api.NewClient(cfg.AccessToken)

			opts := api.InstancesOptions{
				Page: page,
			}

			// If item counter specified, get the item ID first
			if itemCounter > 0 {
				item, err := client.GetItemByCounter(itemCounter)
				if err != nil {
					return fmt.Errorf("getting item #%d: %w", itemCounter, err)
				}
				opts.ItemID = item.ID.Int64()
			}

			instances, err := client.ListInstances(opts)
			if err != nil {
				return err
			}

			// Filter by time if --since specified
			if since != "" {
				sinceTime, parseErr := parseDuration(since)
				if parseErr != nil {
					return fmt.Errorf("invalid --since value: %w", parseErr)
				}

				var filtered []api.Instance
				for _, inst := range instances {
					if inst.Time.After(sinceTime) {
						filtered = append(filtered, inst)
					}
				}
				instances = filtered
			}

			// Apply limit
			if limit > 0 && len(instances) > limit {
				instances = instances[:limit]
			}

			formatter := getFormatter()
			return formatter.FormatInstances(os.Stdout, instances)
		},
	}

	cmd.Flags().IntVar(&itemCounter, "item", 0, "item counter to list occurrences for")
	cmd.Flags().BoolVar(&all, "all", false, "list all project occurrences")
	cmd.Flags().StringVar(&since, "since", "", "filter occurrences since duration")
	cmd.Flags().IntVar(&limit, "limit", 0, "limit number of results")
	cmd.Flags().IntVar(&page, "page", 1, "page number")

	return cmd
}
