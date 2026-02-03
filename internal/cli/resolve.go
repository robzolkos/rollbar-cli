package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/robzolkos/rollbar-cli/internal/api"
)

func newResolveCmd() *cobra.Command {
	var (
		uuid string
	)

	cmd := &cobra.Command{
		Use:   "resolve <counter> [counter...]",
		Short: "Mark items as resolved",
		Long: `Mark one or more items as resolved in Rollbar.

Examples:
  rollbar resolve 123              # Resolve item #123
  rollbar resolve 123 456 789      # Resolve multiple items
  rollbar resolve --uuid abc123    # Resolve by internal ID

Note: This command requires a project access token with write scope.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if uuid == "" && len(args) == 0 {
				return fmt.Errorf("requires at least one item counter or --uuid flag")
			}
			if uuid != "" && len(args) > 0 {
				return fmt.Errorf("cannot specify both --uuid and counter arguments")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				return err
			}

			client := api.NewClient(cfg.AccessToken)

			// Handle UUID mode (single item by internal ID)
			if uuid != "" {
				id, parseErr := strconv.ParseInt(uuid, 10, 64)
				if parseErr != nil {
					return fmt.Errorf("invalid UUID: %w", parseErr)
				}

				item, err := client.UpdateItemStatus(id, "resolved")
				if err != nil {
					return fmt.Errorf("failed to resolve item: %w", err)
				}

				if !quiet {
					fmt.Fprintf(os.Stderr, "Resolved item #%d: %s\n", item.Counter, item.Title)
				}
				return nil
			}

			// Handle counter mode (one or more items by counter)
			var errors []error
			resolved := 0

			for _, arg := range args {
				counter, parseErr := strconv.Atoi(arg)
				if parseErr != nil {
					errors = append(errors, fmt.Errorf("invalid counter %q: %w", arg, parseErr))
					continue
				}

				// First get the item to find its internal ID
				item, err := client.GetItemByCounter(counter)
				if err != nil {
					errors = append(errors, fmt.Errorf("failed to get item #%d: %w", counter, err))
					continue
				}

				// Now resolve it
				_, err = client.UpdateItemStatus(item.ID.Int64(), "resolved")
				if err != nil {
					errors = append(errors, fmt.Errorf("failed to resolve item #%d: %w", counter, err))
					continue
				}

				resolved++
				if !quiet {
					fmt.Fprintf(os.Stderr, "Resolved item #%d: %s\n", item.Counter, item.Title)
				}
			}

			// Report any errors
			if len(errors) > 0 {
				for _, err := range errors {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
				if resolved == 0 {
					return fmt.Errorf("failed to resolve any items")
				}
				return fmt.Errorf("resolved %d item(s), but %d failed", resolved, len(errors))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&uuid, "uuid", "", "resolve item by internal UUID/ID")

	return cmd
}
