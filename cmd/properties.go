package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dotbrains/aptscout/internal/provider"
)

func newPropertiesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "properties",
		Short: "List registered apartment properties",
		Run: func(cmd *cobra.Command, args []string) {
			for id, p := range provider.All {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", id, p.Name())
			}
		},
	}
}
