package ports

import (
	"context"

	"github.com/quest-one/quest-one/internal/domain"
)

// LLMWorker calls an AI provider to generate priority suggestions.
type LLMWorker interface {
	// Prioritize sends a batch of tasks to the LLM and returns an ordered result.
	Prioritize(ctx context.Context, req domain.LLMPrioritizationRequest) (domain.LLMPrioritizationResult, error)
}
