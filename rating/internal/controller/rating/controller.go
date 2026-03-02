package rating

import (
	"context"
	"errors"

	"github.com/tj330/bookapp/rating/pkg/model"
)

var ErrNotFound = errors.New("ratings not found for a record")

type ratingRepository interface {
	Get(ctx context.Context, recordId model.RecordID, recordtype model.RecordType) ([]model.Rating, error)
	Put(ctx context.Context, recordId model.RecordID, recordType model.RecordType, rating *model.Rating) error
}

type Controller struct {
	repo ratingRepository
}

func New(repo ratingRepository) *Controller {
	return &Controller{repo}
}

func (c *Controller) GetAggregatedRating(ctx context.Context, recordId model.RecordID, recordtype model.RecordType) (float64, error) {
	ratings, err := c.repo.Get(ctx, recordId, recordtype)
	if err != nil && errors.Is(err, ErrNotFound) {
		return 0, ErrNotFound
	} else if err != nil {
		return 0, nil
	}

	sum := float64(0)
	for _, r := range ratings {
		sum += float64(r.Value)
	}

	return sum / float64(len(ratings)), nil
}

func (c *Controller) PutRating(ctx context.Context, recordId model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	return c.repo.Put(ctx, recordId, recordType, rating)
}
