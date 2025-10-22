package ports

import (
	"context"

	"github.com/soft-Dev25/real-state-flynect/api/scrappingsvc/internal/core/domain"
)

type StoragePort interface {
	SaveListing(ctx context.Context, listing domain.Listing) error
	GetListings(ctx context.Context) ([]domain.Listing, error)
}
