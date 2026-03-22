package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dotbrains/aptscout/internal/db"
	"github.com/dotbrains/aptscout/internal/display"
	"github.com/dotbrains/aptscout/internal/models"
)

func newListCmd() *cobra.Command {
	var (
		beds        int
		baths       int
		maxPrice    int
		minPrice    int
		plan        string
		renovated   bool
		availableBy string
		sort        string
		jsonOutput  bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available apartments",
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

			f := models.ApartmentFilter{Sort: sort, Order: "asc", Property: propertyFilter()}
			if cmd.Flags().Changed("beds") {
				f.Beds = &beds
			}
			if cmd.Flags().Changed("baths") {
				f.Baths = &baths
			}
			if cmd.Flags().Changed("max-price") {
				f.MaxPrice = &maxPrice
			}
			if cmd.Flags().Changed("min-price") {
				f.MinPrice = &minPrice
			}
			if plan != "" {
				f.Plan = &plan
			}
			if renovated {
				f.Renovated = &renovated
			}
			if availableBy != "" {
				f.AvailableBy = &availableBy
			}

			apts, err := database.ListApartments(f)
			if err != nil {
				return fmt.Errorf("listing apartments: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(apts)
			}

			display.ApartmentTable(cmd.OutOrStdout(), apts)
			return nil
		},
	}

	cmd.Flags().IntVar(&beds, "beds", 0, "filter by bedroom count")
	cmd.Flags().IntVar(&baths, "baths", 0, "filter by bathroom count")
	cmd.Flags().IntVar(&maxPrice, "max-price", 0, "maximum monthly rent")
	cmd.Flags().IntVar(&minPrice, "min-price", 0, "minimum monthly rent")
	cmd.Flags().StringVar(&plan, "plan", "", "filter by floor plan code")
	cmd.Flags().BoolVar(&renovated, "renovated", false, "only renovated units")
	cmd.Flags().StringVar(&availableBy, "available-by", "", "available by date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&sort, "sort", "price", "sort by: price, date, sqft, unit")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	return cmd
}
