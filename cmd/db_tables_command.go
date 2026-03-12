package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDBTablesCommandWithPaths(paths *AppPaths) *cobra.Command {
	return &cobra.Command{
		Use:          "tables",
		Short:        "Show tables and row counts in the database",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, err := paths.DefaultSQLitePath()
			if err != nil {
				return err
			}

			inspector := NewDatabaseInspector(dbPath)
			tables, err := inspector.ListTables()
			if err != nil {
				return err
			}

			for _, table := range tables {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%d\n", table.Name, table.RowCount); err != nil {
					return fmt.Errorf("write database tables output: %w", err)
				}
			}

			return nil
		},
	}
}
