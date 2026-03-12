// Package cli implements the cobra command tree.
package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/quest-one/quest-one/internal/application"
)

// rootCmd is the base command.
var rootCmd *cobra.Command

// Build constructs and returns the root cobra command with all subcommands registered.
func Build(app *application.App) *cobra.Command {
	rootCmd = &cobra.Command{
		Use:   "quest-one",
		Short: "Local task management with multi-source sync and AI prioritization",
		Long: `quest-one is a local-first task manager that aggregates work items from
Redmine, Slack, Google Calendar, NotePM, and Google Drive, then helps you
decide what to work on next using configurable priority scoring and LLM assistance.`,
	}

	rootCmd.AddCommand(
		serveCmd(app),
		mcpCmd(app),
		setupCmd(app),
		nextCmd(app),
		addCmd(app),
		listCmd(app),
		completeCmd(app),
		cancelCmd(app),
		promoteCmd(app),
		memoCmd(app),
		candidatesCmd(app),
		searchCmd(app),
		syncCmd(app),
		settingsCmd(app),
		integrationsCmd(app),
		logsCmd(app),
		voiceCmd(app),
	)

	return rootCmd
}

// Execute runs the CLI. Call from main().
func Execute(ctx context.Context, app *application.App) error {
	return Build(app).ExecuteContext(ctx)
}
