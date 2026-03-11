package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newBatchCommand() *cobra.Command {
	return newBatchCommandWithPaths(NewDefaultAppPaths())
}

func newBatchCommandWithPaths(paths *AppPaths) *cobra.Command {
	var dbPath string
	var fromPage int
	var toPage int

	command := &cobra.Command{
		Use:          "batch <base-url>",
		Short:        "Scrape multiple paginated pages into SQLite",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("to") {
				return nil
			}

			if fromPage < 1 {
				return fmt.Errorf("--from must be a positive integer greater than or equal to 1")
			}

			if toPage < 1 {
				return fmt.Errorf("--to must be a positive integer greater than or equal to 1")
			}

			if toPage < fromPage {
				return fmt.Errorf("--to must be greater than or equal to --from")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedPath, err := resolveSQLitePath(paths, dbPath)
			if err != nil {
				return err
			}

			runner := NewBatchRunner(cmd.OutOrStdout())
			return runner.Run(args[0], resolvedPath, fromPage, toPage)
		},
	}

	command.Flags().IntVar(&fromPage, "from", 1, "Starting page number (inclusive)")
	command.Flags().IntVar(&toPage, "to", 0, "Ending page number (inclusive)")
	addSQLitePathFlag(command, paths, &dbPath)

	if err := command.MarkFlagRequired("to"); err != nil {
		panic(err)
	}

	return command
}

func init() {
	rootCmd.AddCommand(newBatchCommand())
}
