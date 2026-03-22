package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dotbrains/aptscout/internal/db"
	"github.com/dotbrains/aptscout/internal/display"
	"github.com/dotbrains/aptscout/internal/models"
	"github.com/dotbrains/aptscout/internal/provider"
)

func newHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history <unit_number>",
		Short: "Show price history for a unit",
		Args:  cobra.ExactArgs(1),
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

			unitNumber := args[0]
			// Try to find the unit across all properties, or scoped to --property.
			var apt *models.Apartment
			if flagProperty != "" {
				apt, _ = database.GetApartment(flagProperty, unitNumber)
			} else {
				// Search all properties for this unit.
				for _, pid := range provider.IDs() {
					a, e := database.GetApartment(pid, unitNumber)
					if e == nil {
						apt = a
						break
					}
				}
			}
			if apt == nil {
				return fmt.Errorf("unit #%s not found", unitNumber)
			}

			records, err := database.GetPriceHistory(apt.Property, unitNumber)
			if err != nil {
				return fmt.Errorf("fetching price history: %w", err)
			}

			if len(records) == 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No price history for unit #%s\n", unitNumber)
				return nil
			}

			display.PriceHistoryTable(cmd.OutOrStdout(), apt, records)
			return nil
		},
	}
}
