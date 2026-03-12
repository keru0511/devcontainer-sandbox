// Package events provides a simple in-process event publisher.
package events

import (
	"context"
	"log/slog"

	"github.com/quest-one/quest-one/internal/domain"
	"github.com/quest-one/quest-one/internal/ports"
)

// LogPublisher logs every domain event to slog. Suitable for development;
// replace with a fan-out publisher in production if webhooks are needed.
type LogPublisher struct {
	log *slog.Logger
}

var _ ports.EventPublisher = (*LogPublisher)(nil)

// NewLogPublisher returns a publisher that logs events at Debug level.
func NewLogPublisher(log *slog.Logger) *LogPublisher {
	return &LogPublisher{log: log}
}

func (p *LogPublisher) Publish(_ context.Context, events ...domain.DomainEvent) error {
	for _, e := range events {
		p.log.Debug("domain_event", slog.Any("event", e))
	}
	return nil
}
