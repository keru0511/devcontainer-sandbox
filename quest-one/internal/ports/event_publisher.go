package ports

import (
	"context"

	"github.com/quest-one/quest-one/internal/domain"
)

// EventPublisher dispatches domain events after a transaction commits.
// Implementations may fan-out to multiple handlers (logging, webhooks, sync triggers).
type EventPublisher interface {
	Publish(ctx context.Context, events ...domain.DomainEvent) error
}
