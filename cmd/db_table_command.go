package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDBTableCommandWithPaths(paths *AppPaths) *cobra.Command {
	return &cobra.Command{
		Use:          "table <table-name>",
		Short:        "Show the schema for a database table",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, err := paths.DefaultSQLitePath()
			if err != nil {
				return err
			}

			inspector := NewDatabaseInspector(dbPath)
			columns, err := inspector.DescribeTable(args[0])
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "name\ttype\tnotnull\tdefault\tpk"); err != nil {
				return fmt.Errorf("write table schema header: %w", err)
			}

			for _, column := range columns {
				if _, err := fmt.Fprintf(
					cmd.OutOrStdout(),
					"%s\t%s\t%t\t%s\t%t\n",
					column.Name,
					column.Type,
					column.NotNull,
					column.DefaultValue,
					column.PrimaryKey,
				); err != nil {
					return fmt.Errorf("write table schema output: %w", err)
				}
			}

			return nil
		},
	}
}
