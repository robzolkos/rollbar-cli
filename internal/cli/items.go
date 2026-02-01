package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/robzolkos/rollbar-cli/internal/api"
)

func newItemsCmd() *cobra.Command {
	var (
		status string
		level  string
		env    string
		query  string
		since  string
		from   string
		to     string
		sortBy string
		page   int
		limit  int
	)

	cmd := &cobra.Command{
		Use:   "items",
		Short: "List items (errors)",
		Long: `List items (errors) from Rollbar with various filters.

Examples:
  rollbar items                              # List active items
  rollbar items --status resolved            # List resolved items
  rollbar items --level error,critical       # Filter by level
  rollbar items --env production             # Filter by environment
  rollbar items --since "8 hours ago"        # Items from last 8 hours
  rollbar items --since 24h                  # Items from last 24 hours
  rollbar items --query "TypeError"          # Search by title
  rollbar items --sort occurrences           # Sort by occurrence count
  rollbar items --ai                         # Token-efficient output for AI`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				return err
			}

			client := api.NewClient(cfg.AccessToken)

			opts := api.ItemsOptions{
				Status:      status,
				Level:       level,
				Environment: env,
				Query:       query,
				Page:        page,
			}

			// Parse time filters
			if since != "" {
				t, err := parseDuration(since)
				if err != nil {
					return fmt.Errorf("invalid --since value: %w", err)
				}
				opts.DateFrom = t
			}
			if from != "" {
				t, err := parseTimeArg(from)
				if err != nil {
					return fmt.Errorf("invalid --from value: %w", err)
				}
				opts.DateFrom = t
			}
			if to != "" {
				t, err := parseTimeArg(to)
				if err != nil {
					return fmt.Errorf("invalid --to value: %w", err)
				}
				opts.DateTo = t
			}

			// Use default environment from config if not specified
			if opts.Environment == "" && cfg.DefaultEnvironment != "" {
				opts.Environment = cfg.DefaultEnvironment
			}

			items, _, err := client.ListItems(opts)
			if err != nil {
				return err
			}

			// Apply sorting
			items = sortItems(items, sortBy)

			// Apply limit
			if limit > 0 && len(items) > limit {
				items = items[:limit]
			}

			formatter := getFormatter()
			return formatter.FormatItems(os.Stdout, items)
		},
	}

	cmd.Flags().StringVar(&status, "status", "active", "filter by status: active, resolved, muted, any")
	cmd.Flags().StringVar(&level, "level", "", "filter by level: debug, info, warning, error, critical (comma-separated)")
	cmd.Flags().StringVar(&env, "env", "", "filter by environment")
	cmd.Flags().StringVar(&query, "query", "", "text search in item titles")
	cmd.Flags().StringVar(&since, "since", "", "filter items with occurrences since duration (e.g., '8 hours ago', '24h', '7 days')")
	cmd.Flags().StringVar(&from, "from", "", "filter items from datetime (ISO 8601)")
	cmd.Flags().StringVar(&to, "to", "", "filter items until datetime (ISO 8601)")
	cmd.Flags().StringVar(&sortBy, "sort", "recent", "sort by: recent, occurrences, first-seen, level")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	cmd.Flags().IntVar(&limit, "limit", 0, "limit number of results (0 = no limit)")

	return cmd
}

func sortItems(items []api.Item, sortBy string) []api.Item {
	switch strings.ToLower(sortBy) {
	case "occurrences":
		sort.Slice(items, func(i, j int) bool {
			return items[i].TotalOccurrences > items[j].TotalOccurrences
		})
	case "first-seen":
		sort.Slice(items, func(i, j int) bool {
			return items[i].FirstOccurrenceTimestamp > items[j].FirstOccurrenceTimestamp
		})
	case "level":
		sort.Slice(items, func(i, j int) bool {
			return items[i].Level.Int() > items[j].Level.Int() // Higher level = more severe
		})
	case "recent":
		fallthrough
	default:
		sort.Slice(items, func(i, j int) bool {
			return items[i].LastOccurrenceTimestamp > items[j].LastOccurrenceTimestamp
		})
	}
	return items
}
