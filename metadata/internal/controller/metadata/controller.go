package metadata

import (
	"context"
	"errors"

	"github.com/tj330/bookapp/metadata/internal/repository"
	"github.com/tj330/bookapp/metadata/pkg/model"
)

// Custom error when the metadata for a book is not found.
var ErrNotFound = errors.New("not found")

// metadataRepository defines the storage interface for retrieving
// and persisting metadata records.
type metadataRepository interface {
	Get(ctx context.Context, id string) (*model.Metadata, error)
	Put(ctx context.Context, id string, metadata *model.Metadata) error
}

// Controller handles the business logic for metadata service.
type Controller struct {
	repo metadataRepository
}

// New returns a new metadata controller with the metadataRepository.
func New(repo metadataRepository) *Controller {
	return &Controller{repo}
}

// Get finds the metadata by its id.
func (c *Controller) Get(ctx context.Context, id string) (*model.Metadata, error) {
	res, err := c.repo.Get(ctx, id)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return res, err
}

// Put saves new metadata.
func (c *Controller) Put(ctx context.Context, m *model.Metadata) error {
	return c.repo.Put(ctx, m.ID, m)
}
