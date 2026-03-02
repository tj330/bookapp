package book

import (
	"context"
	"errors"

	"github.com/tj330/bookapp/book/internal/gateway"
	"github.com/tj330/bookapp/book/pkg/model"
	metadatamodel "github.com/tj330/bookapp/metadata/pkg/model"
	ratingmodel "github.com/tj330/bookapp/rating/pkg/model"
)

var ErrNotFound = errors.New("book metadata not found")

type ratingGateway interface {
	GetAggregatedRating(ctx context.Context, recordId ratingmodel.RecordID, recordType ratingmodel.RecordType) (float64, error)
	PutRating(ctx context.Context, recordId ratingmodel.RecordID, recordType ratingmodel.RecordType, rating *ratingmodel.Rating) error
}

type metadatGateway interface {
	Get(ctx context.Context, id string) (*metadatamodel.Metadata, error)
}

type Controller struct {
	ratingGateway
	metadatGateway
}

func New(ratingGateway ratingGateway, metadatGateway metadatGateway) *Controller {
	return &Controller{ratingGateway, metadatGateway}
}

func (c *Controller) Get(ctx context.Context, id string) (*model.BookDetails, error) {
	metadata, err := c.metadatGateway.Get(ctx, id)
	if err != nil && errors.Is(gateway.ErrNotFound, err) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	details := &model.BookDetails{Metadata: *metadata}
	rating, err := c.ratingGateway.GetAggregatedRating(ctx, ratingmodel.RecordID(id), ratingmodel.RecordTypeBook)
	if errors.Is(gateway.ErrNotFound, err) {

	} else if err != nil {
		return nil, err
	} else {
		details.Rating = &rating
	}
	return details, err
}
