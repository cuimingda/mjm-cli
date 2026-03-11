package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDBPathCommandWithPaths(paths *AppPaths) *cobra.Command {
	return &cobra.Command{
		Use:          "path",
		Short:        "Show the database file path",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, err := paths.DefaultSQLitePath()
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintln(cmd.OutOrStdout(), dbPath); err != nil {
				return fmt.Errorf("write database path output: %w", err)
			}

			return nil
		},
	}
}
