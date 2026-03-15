package cli

import (
	"github.com/spf13/cobra"

	"github.com/quest-one/quest-one/internal/application"
	"github.com/quest-one/quest-one/internal/mcp"
)

func mcpCmd(app *application.App) *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP (Model Context Protocol) stdio server",
		Long: `Starts a JSON-RPC 2.0 server on stdin/stdout that implements the
Model Context Protocol, allowing AI assistants (Claude, etc.) to interact
with quest-one as a tool provider.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := mcp.New(app, app.Log)
			return s.Run(cmd.Context())
		},
	}
}
