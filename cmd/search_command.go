package cmd

import "github.com/spf13/cobra"

func newSearchCommand() *cobra.Command {
	return newSearchCommandWithPaths(NewDefaultAppPaths())
}

func newSearchCommandWithPaths(paths *AppPaths) *cobra.Command {
	var dbPath string
	var sortByTitle bool

	command := &cobra.Command{
		Use:          "search TERM [TERM...]",
		Short:        "Search scraped entry titles with SQLite FTS5",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedPath, err := resolveSQLitePath(paths, dbPath)
			if err != nil {
				return err
			}

			runner := NewSearchRunner(cmd.OutOrStdout())
			return runner.Run(resolvedPath, args, SearchOptions{
				SortByTitle: sortByTitle,
			})
		},
	}

	addSQLitePathFlag(command, paths, &dbPath)
	command.Flags().BoolVar(&sortByTitle, "sort-by-title", false, "Sort search results by title instead of href")

	return command
}

func init() {
	rootCmd.AddCommand(newSearchCommand())
}
