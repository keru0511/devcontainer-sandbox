package ports

import (
	"context"

	"github.com/quest-one/quest-one/internal/domain"
)

// IntegrationRepository persists Integration configurations.
type IntegrationRepository interface {
	Save(ctx context.Context, integration domain.Integration) error
	FindByID(ctx context.Context, id domain.IntegrationID) (domain.Integration, error)
	FindAll(ctx context.Context) ([]domain.Integration, error)
	FindEnabled(ctx context.Context) ([]domain.Integration, error)
	Delete(ctx context.Context, id domain.IntegrationID) error
}
