package repositories

import (
	"context"
	"sync"

	"github.com/soft-Dev25/real-state-flynect/api/scrappingsvc/internal/core/domain"
)

type MemoryRepo struct {
	mu       sync.Mutex
	listings []domain.Listing
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{}
}

func (r *MemoryRepo) SaveListing(ctx context.Context, listing domain.Listing) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.listings = append(r.listings, listing)
	return nil
}

func (r *MemoryRepo) GetListings(ctx context.Context) ([]domain.Listing, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]domain.Listing(nil), r.listings...), nil
}
