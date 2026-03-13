package cmd

import "github.com/spf13/cobra"

func newSearchCommand() *cobra.Command {
	return newSearchCommandWithPaths(NewDefaultAppPaths())
}

func newSearchCommandWithPaths(paths *AppPaths) *cobra.Command {
	var dbPath string

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
			return runner.Run(resolvedPath, args)
		},
	}

	addSQLitePathFlag(command, paths, &dbPath)

	return command
}

func init() {
	rootCmd.AddCommand(newSearchCommand())
}
