// Package application contains use cases. Each use case is one exported function
// that executes within a single logical transaction.
package application

import (
	"log/slog"

	"github.com/quest-one/quest-one/internal/ports"
)

// App aggregates all ports required by the use cases.
// It is constructed once at startup and injected into HTTP handlers and CLI commands.
type App struct {
	Tasks        ports.TaskRepository
	SourceItems  ports.SourceItemRepository
	Integrations ports.IntegrationRepository
	Settings     ports.SettingsRepository
	Events       ports.EventPublisher
	Priority     ports.PriorityEngine
	LLM          ports.LLMWorker // may be nil if LLM is disabled
	Search       ports.SearchIndex
	Adapters     []ports.SourceAdapter // registered source adapters for sync
	Log          *slog.Logger
}
