package cmd

import (
	"github.com/spf13/cobra"
)

const defaultScrapeDBPath = "mjm.sqlite"

func newScrapeCommand() *cobra.Command {
	var dbPath string

	command := &cobra.Command{
		Use:          "scrape <url>",
		Short:        "Scrape entry links into SQLite",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := NewScrapeRunner(cmd.OutOrStdout())
			return runner.Run(args[0], dbPath)
		},
	}

	command.Flags().StringVar(&dbPath, "db", defaultScrapeDBPath, "SQLite database path")

	return command
}

func init() {
	rootCmd.AddCommand(newScrapeCommand())
}
