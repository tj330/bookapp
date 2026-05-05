package metadata

import (
	"context"
	"errors"

	"github.com/tj330/bookapp/metadata/internal/repository"
	"github.com/tj330/bookapp/metadata/pkg/model"
)

var ErrNotFound = errors.New("not found")

type metadataRepository interface {
	Get(ctx context.Context, id string) (*model.Metadata, error)
	Put(ctx context.Context, id string, metadata *model.Metadata) error
}

type Controller struct {
	repo metadataRepository
}

func New(repo metadataRepository) *Controller {
	return &Controller{repo}
}

func (c *Controller) Get(ctx context.Context, id string) (*model.Metadata, error) {
	res, err := c.repo.Get(ctx, id)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return res, err
}

func (c *Controller) Put(ctx context.Context, m *model.Metadata) error {
	return c.repo.Put(ctx, m.ID, m)
}
