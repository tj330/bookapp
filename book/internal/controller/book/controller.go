package book

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/tj330/bookapp/book/internal/gateway"
	"github.com/tj330/bookapp/book/pkg/model"
	metadatamodel "github.com/tj330/bookapp/metadata/pkg/model"
	ratingmodel "github.com/tj330/bookapp/rating/pkg/model"
)

var ErrNotFound = errors.New("book metadata not found")

type ratingGateway interface {
	GetAggregatedRating(ctx context.Context, recordId ratingmodel.RecordID, recordType ratingmodel.RecordType) (float64, error)
}

type metadataGateway interface {
	Get(ctx context.Context, id string) (*metadatamodel.Metadata, error)
}

type Controller struct {
	ratingGateway
	metadataGateway
}

func New(ratingGateway ratingGateway, metadatGateway metadataGateway) *Controller {
	return &Controller{ratingGateway, metadatGateway}
}

func (c *Controller) Get(ctx context.Context, id string) (*model.BookDetails, error) {
	var wg sync.WaitGroup
	wg.Add(2)
	var metadata *metadatamodel.Metadata
	var getMetadaErr error
	var rating float64
	var getRatingErr error

	go func() {
		defer wg.Done()
		metadata, getMetadaErr = c.metadataGateway.Get(ctx, id)
	}()
	go func() {
		defer wg.Done()
		rating, getRatingErr = c.ratingGateway.GetAggregatedRating(ctx, ratingmodel.RecordID(id), ratingmodel.RecordTypeBook)
	}()
	wg.Wait()

	if getMetadaErr != nil {
		if errors.Is(getMetadaErr, gateway.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, getMetadaErr
	}

	if metadata == nil {
		return nil, fmt.Errorf("metadata gateway returned no data and no error")
	}

	details := &model.BookDetails{Metadata: *metadata}

	if getRatingErr != nil {
		if errors.Is(getRatingErr, gateway.ErrNotFound) {
			return details, nil
		}
		return nil, getRatingErr
	}
	details.Rating = &rating

	return details, nil
}
