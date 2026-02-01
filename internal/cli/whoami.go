package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/robzolkos/rollbar-cli/internal/api"
)

func newWhoamiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Test authentication and show project info",
		Long: `Verify that your access token is valid and show the associated project.

Examples:
  rollbar whoami    # Check authentication`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				return err
			}

			client := api.NewClient(cfg.AccessToken)
			info, err := client.GetProjectInfo()
			if err != nil {
				return err
			}

			formatter := getFormatter()
			return formatter.FormatProjectInfo(os.Stdout, info)
		},
	}

	return cmd
}
