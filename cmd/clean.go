package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dotbrains/aptscout/internal/db"
)

func newCleanCmd() *cobra.Command {
	var days int
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove stale apartment data",
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

			if dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "→ Would remove apartments not seen in %d days (dry run)\n", days)
				return nil
			}

			removed, err := database.CleanStale(days)
			if err != nil {
				return fmt.Errorf("cleaning stale data: %w", err)
			}

			if removed == 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✓ No stale records to clean.\n")
			} else {
				n := int(removed)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "→ Removing %d apartment%s not seen in %d days...\n", n, plural(n), days)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✓ Cleaned %d stale record%s.\n", n, plural(n))
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 30, "remove apartments not seen in N days")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be removed")

	return cmd
}
