package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	defaultDeveloperID = "mingda.dev"
	defaultAppName     = "mjm"
)

func newScrapeCommand() *cobra.Command {
	return newScrapeCommandWithPaths(NewAppPaths(defaultDeveloperID, defaultAppName))
}

func newScrapeCommandWithPaths(paths *AppPaths) *cobra.Command {
	var dbPath string
	defaultDBPath, _ := paths.DefaultSQLitePath()

	command := &cobra.Command{
		Use:          "scrape <url>",
		Short:        "Scrape entry links into SQLite",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				resolvedPath, err := paths.DefaultSQLitePath()
				if err != nil {
					return err
				}
				dbPath = resolvedPath
			}

			runner := NewScrapeRunner(cmd.OutOrStdout())
			return runner.Run(args[0], dbPath)
		},
	}

	command.Flags().StringVar(&dbPath, "db", defaultDBPath, fmt.Sprintf("SQLite database path (default %s)", defaultDBPath))

	return command
}

func init() {
	rootCmd.AddCommand(newScrapeCommand())
}
