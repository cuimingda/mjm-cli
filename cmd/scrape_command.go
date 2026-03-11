package cmd

import "github.com/spf13/cobra"

func newScrapeCommand() *cobra.Command {
	return newScrapeCommandWithPaths(NewDefaultAppPaths())
}

func newScrapeCommandWithPaths(paths *AppPaths) *cobra.Command {
	var dbPath string

	command := &cobra.Command{
		Use:          "scrape <url>",
		Short:        "Scrape entry links into SQLite",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedPath, err := resolveSQLitePath(paths, dbPath)
			if err != nil {
				return err
			}

			runner := NewScrapeRunner(cmd.OutOrStdout())
			return runner.Run(args[0], resolvedPath)
		},
	}

	addSQLitePathFlag(command, paths, &dbPath)

	return command
}

func init() {
	rootCmd.AddCommand(newScrapeCommand())
}
