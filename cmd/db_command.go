package cmd

import "github.com/spf13/cobra"

func newDBCommand() *cobra.Command {
	return newDBCommandWithPaths(NewDefaultAppPaths())
}

func newDBCommandWithPaths(paths *AppPaths) *cobra.Command {
	command := &cobra.Command{
		Use:   "db",
		Short: "Database utilities",
	}

	command.AddCommand(newDBPathCommandWithPaths(paths))
	command.AddCommand(newDBTablesCommandWithPaths(paths))

	return command
}

func init() {
	rootCmd.AddCommand(newDBCommand())
}
