package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagDB       string
	flagProperty string
)

func newRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "aptscout",
		Short: "Apartment availability tracker",
		Long:  "Scrape, store, and browse apartment availability across multiple properties.",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		Version: version,
	}

	root.SetVersionTemplate(fmt.Sprintf("aptscout version %s\n", version))

	root.PersistentFlags().StringVar(&flagDB, "db", "", "override database path")
	root.PersistentFlags().StringVar(&flagProperty, "property", "", "filter by property (e.g. desert-club, hideaway)")

	root.AddCommand(newScrapeCmd(version))
	root.AddCommand(newListCmd())
	root.AddCommand(newHistoryCmd())
	root.AddCommand(newStatsCmd())
	root.AddCommand(newServeCmd(version))
	root.AddCommand(newCleanCmd())
	root.AddCommand(newPropertiesCmd())

	return root
}

func propertyFilter() *string {
	if flagProperty == "" {
		return nil
	}
	return &flagProperty
}

// Execute runs the root command.
func Execute(version string) error {
	return newRootCmd(version).Execute()
}
