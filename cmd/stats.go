package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dotbrains/aptscout/internal/db"
	"github.com/dotbrains/aptscout/internal/display"
)

func newStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show summary statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := flagDB
			if dbPath == "" {
				dbPath = db.DefaultPath()
			}

			database, err := db.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening database: %w", err)
			}
			defer func() { _ = database.Close() }()

			stats, err := database.GetStats(propertyFilter())
			if err != nil {
				return fmt.Errorf("fetching stats: %w", err)
			}

			display.StatsDisplay(cmd.OutOrStdout(), stats)
			return nil
		},
	}
}
