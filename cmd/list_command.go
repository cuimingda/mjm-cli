package cmd

import "github.com/spf13/cobra"

func newListCommand() *cobra.Command {
	return newListCommandWithPaths(NewDefaultAppPaths())
}

func newListCommandWithPaths(paths *AppPaths) *cobra.Command {
	var dbPath string

	command := &cobra.Command{
		Use:          "list",
		Short:        "List scraped entries from SQLite",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedPath, err := resolveSQLitePath(paths, dbPath)
			if err != nil {
				return err
			}

			runner := NewListRunner(cmd.OutOrStdout())
			return runner.Run(resolvedPath)
		},
	}

	addSQLitePathFlag(command, paths, &dbPath)

	return command
}

func init() {
	rootCmd.AddCommand(newListCommand())
}
