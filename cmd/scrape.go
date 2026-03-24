package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dotbrains/aptscout/internal/db"
	"github.com/dotbrains/aptscout/internal/models"
	"github.com/dotbrains/aptscout/internal/provider"
	"github.com/dotbrains/aptscout/internal/scraper"
	"github.com/dotbrains/aptscout/internal/spinner"
)

func newScrapeCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "scrape",
		Short: "Scrape floor plans and update the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScrape(cmd, version)
		},
	}
}

func runScrape(cmd *cobra.Command, version string) error {
	dbPath := flagDB
	if dbPath == "" {
		dbPath = db.DefaultPath()
	}

	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() { _ = database.Close() }()

	// Determine which providers to scrape.
	providers := resolveProviders()
	if len(providers) == 0 {
		return fmt.Errorf("unknown property: %s", flagProperty)
	}

	w := cmd.OutOrStdout()
	s := scraper.New(database, w)
	var totalFound, totalNew, totalChanged, totalRemoved int

	for i, prov := range providers {
		_, _ = fmt.Fprintf(w, "\n[%d/%d] %s\n", i+1, len(providers), prov.Name())

		sp := spinner.New(w, "  Fetching floor plans and units...")
		sp.Start()

		result, err := s.RunProvider(context.Background(), prov)
		sp.Stop()

		if err != nil {
			_, _ = fmt.Fprintf(w, "  ✗ %s: %v\n", prov.Name(), err)
			continue
		}
		_, _ = fmt.Fprintf(w, "  ✓ %d plans, %d units", result.FloorPlans, result.UnitsFound)
		if result.UnitsNew > 0 {
			_, _ = fmt.Fprintf(w, " (%d new)", result.UnitsNew)
		}
		if result.UnitsChanged > 0 {
			_, _ = fmt.Fprintf(w, " (%d changed)", result.UnitsChanged)
		}
		_, _ = fmt.Fprintln(w)

		totalFound += result.UnitsFound
		totalNew += result.UnitsNew
		totalChanged += result.UnitsChanged
		totalRemoved += result.UnitsRemoved
	}

	_, _ = fmt.Fprintf(w, "\n✓ Scrape complete.\n")
	parts := []string{fmt.Sprintf("%d unit%s available", totalFound, plural(totalFound))}
	if totalNew > 0 {
		parts = append(parts, fmt.Sprintf("%d new", totalNew))
	}
	if totalChanged > 0 {
		parts = append(parts, fmt.Sprintf("%d price changed", totalChanged))
	}
	if totalRemoved > 0 {
		parts = append(parts, fmt.Sprintf("%d no longer available", totalRemoved))
	}
	if len(parts) > 1 {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\u2192 %s (%s)\n", parts[0], joinParts(parts[1:]))
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\u2192 %s\n", parts[0])
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\u2192 Database: %s\n", dbPath)
	return nil
}

func resolveProviders() []models.Provider {
	if flagProperty != "" {
		p := provider.Get(flagProperty)
		if p == nil {
			return nil
		}
		return []models.Provider{p}
	}
	var all []models.Provider
	for _, p := range provider.All {
		all = append(all, p)
	}
	return all
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func joinParts(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
