package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/dotbrains/aptscout/internal/db"
	"github.com/dotbrains/aptscout/internal/server"
)

func newServeCmd(version string) *cobra.Command {
	var port int
	var open bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start a local web UI to browse apartments",
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

			addr := fmt.Sprintf(":%d", port)
			url := fmt.Sprintf("http://localhost:%d", port)

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "→ Serving at %s\n", url)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "→ Database: %s\n", dbPath)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "→ Press Ctrl+C to stop\n")

			if open {
				openBrowser(url)
			}

			srv := server.New(database)
			return srv.ListenAndServe(addr)
		},
	}

	cmd.Flags().IntVar(&port, "port", 8700, "port to serve on")
	cmd.Flags().BoolVar(&open, "open", false, "open browser automatically")

	return cmd
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return
	}
	_ = cmd.Start()
}
