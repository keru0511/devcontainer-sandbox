package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/quest-one/quest-one/internal/application"
	"github.com/quest-one/quest-one/internal/server"
)

func serveCmd(app *application.App) *cobra.Command {
	var port int

	c := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP server and web UI",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if port == 0 {
				settings, err := app.GetSettings(cmd.Context())
				if err == nil {
					port = settings.ServerPort
				}
				if port == 0 {
					port = 7890
				}
			}

			s := server.New(app, server.Config{Port: port, Log: app.Log})

			fmt.Fprintf(cmd.OutOrStdout(), "Serving on http://localhost:%d\n", port)
			return s.Start(cmd.Context())
		},
	}
	c.Flags().IntVarP(&port, "port", "p", 0, "Override server port")
	return c
}

func setupCmd(app *application.App) *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Interactive first-run setup wizard",
		RunE: func(cmd *cobra.Command, _ []string) error {
			settings, err := app.GetSettings(cmd.Context())
			w := cmd.OutOrStdout()
			if err != nil {
				fmt.Fprintf(w, "Warning: could not load existing settings (%v), using defaults.\n", err)
			}

			fmt.Fprintln(w, "=== quest-one setup ===")
			fmt.Fprintf(w, "Data directory [%s]: ", settings.DataDir)

			// Read from stdin (simplified; a full TUI would use huh or bubbletea)
			var input string
			fmt.Fscanln(cmd.InOrStdin(), &input)
			if input != "" {
				settings.DataDir = input
			}

			fmt.Fprintf(w, "Server port [%d]: ", settings.ServerPort)
			fmt.Fscanln(cmd.InOrStdin(), &input)
			if input != "" {
				if p, err := strconv.Atoi(input); err == nil {
					settings.ServerPort = p
				}
			}

			if err := app.UpdateSettings(cmd.Context(), settings); err != nil {
				return fmt.Errorf("setup: save settings: %w", err)
			}
			fmt.Fprintln(w, "Settings saved. Run `quest-one serve` to start.")
			return nil
		},
	}
}

func syncCmd(app *application.App) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync all enabled integrations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			results, err := app.SyncAll(cmd.Context(), nil)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			for _, r := range results {
				fmt.Fprintf(w, "%s: +%d updated:%d errors:%d\n",
					r.SourceType, r.Created, r.Updated, len(r.Errors))
			}
			if len(results) == 0 {
				fmt.Fprintln(w, "No enabled integrations to sync.")
			}
			return nil
		},
	}
}
