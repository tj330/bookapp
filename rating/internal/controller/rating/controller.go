package rating

import (
	"context"
	"errors"
	"fmt"

	"github.com/tj330/bookapp/rating/pkg/model"
)

// Custom error if rating is not found for a record.
var ErrNotFound = errors.New("ratings not found for a record")

// ratingRepository interface containing methods to Get and Put and rating.
type ratingRepository interface {
	Get(ctx context.Context, recordId model.RecordID, recordtype model.RecordType) ([]model.Rating, error)
	Put(ctx context.Context, recordId model.RecordID, recordType model.RecordType, rating *model.Rating) error
}

// Controller handle the business logic for rating service.
type Controller struct {
	repo ratingRepository
	// For asynchronous ingestion of rating events.
	ingester ratingIngester
}

// New returns a new controller with
func New(repo ratingRepository, ingester ratingIngester) *Controller {
	return &Controller{repo, ingester}
}

// GetAggregatedRating returns the average rating from all of the ratings for a specific book.
func (c *Controller) GetAggregatedRating(ctx context.Context, recordId model.RecordID, recordtype model.RecordType) (float64, error) {
	ratings, err := c.repo.Get(ctx, recordId, recordtype)
	if err != nil {
		return 0, err
	}
	if len(ratings) == 0 {
		return 0, ErrNotFound
	}

	sum := float64(0)
	for _, r := range ratings {
		sum += float64(r.Value)
	}

	return sum / float64(len(ratings)), nil
}

// PutRating is used to insert a new rating for a book.
func (c *Controller) PutRating(ctx context.Context, recordId model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	return c.repo.Put(ctx, recordId, recordType, rating)
}

type ratingIngester interface {
	Ingest(ctx context.Context) (chan model.RatingEvent, error)
}

func (s *Controller) StartIngestion(ctx context.Context) error {
	ch, err := s.ingester.Ingest(ctx)
	if err != nil {
		return err
	}

	for e := range ch {
		fmt.Printf("Consumed a message: %v\n", e)
		if err := s.PutRating(ctx, e.RecordID, e.RecordType, &model.Rating{}); err != nil {
			return err
		}
	}

	return nil
}
