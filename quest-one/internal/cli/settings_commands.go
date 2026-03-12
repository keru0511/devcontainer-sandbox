package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/quest-one/quest-one/internal/application"
	"github.com/quest-one/quest-one/internal/domain"
)

func settingsCmd(app *application.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "settings",
		Short: "View or update settings",
	}

	c.AddCommand(
		&cobra.Command{
			Use:   "get",
			Short: "Show current settings",
			RunE: func(cmd *cobra.Command, _ []string) error {
				s, err := app.GetSettings(cmd.Context())
				if err != nil {
					return err
				}
				b, _ := json.MarshalIndent(s, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(b))
				return nil
			},
		},
		&cobra.Command{
			Use:   "set-port <port>",
			Short: "Change the server port",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				s, err := app.GetSettings(cmd.Context())
				if err != nil {
					return err
				}
				var port int
				if _, err := fmt.Sscan(args[0], &port); err != nil {
					return fmt.Errorf("%w: invalid port", domain.ErrInvalidInput)
				}
				s.ServerPort = port
				return app.UpdateSettings(cmd.Context(), s)
			},
		},
	)

	return c
}

func integrationsCmd(app *application.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "integrations",
		Short: "Manage external integrations",
	}

	c.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List all integrations",
			RunE: func(cmd *cobra.Command, _ []string) error {
				integrations, err := app.ListIntegrations(cmd.Context())
				if err != nil {
					return err
				}
				if len(integrations) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "No integrations configured.")
					return nil
				}
				for _, i := range integrations {
					status := "disabled"
					if i.Enabled {
						status = "enabled"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s  %-10s  %-8s  %s\n",
						i.ID, i.Provider, status, i.Name)
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "enable <id>",
			Short: "Enable an integration",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				_, err := app.EnableIntegration(cmd.Context(), domain.IntegrationID(args[0]))
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Integration enabled.")
				return nil
			},
		},
		&cobra.Command{
			Use:   "disable <id>",
			Short: "Disable an integration",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				_, err := app.DisableIntegration(cmd.Context(), domain.IntegrationID(args[0]))
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Integration disabled.")
				return nil
			},
		},
	)

	return c
}
