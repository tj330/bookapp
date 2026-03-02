package metadata

import (
	"context"
	"errors"

	"github.com/tj330/bookapp/metadata/pkg/model"
)

var ErrNotFound = errors.New("not found")

type metadataRepository interface {
	Get(ctx context.Context, id string) (*model.Metadata, error)
}

type Controller struct {
	repo metadataRepository
}

func New(repo metadataRepository) *Controller {
	return &Controller{repo}
}

func (c *Controller) Get(ctx context.Context, id string) (*model.Metadata, error) {
	res, err := c.repo.Get(ctx, id)
	if err != nil && errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}

	return res, nil
}
