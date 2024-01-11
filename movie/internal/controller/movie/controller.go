package movie

import (
	"context"
	"errors"

	metadataModel "movieexample.com/metadata/pkg/model"
	"movieexample.com/movie/internal/gateway"
	"movieexample.com/movie/pkg/model"
	ratingModel "movieexample.com/rating/pkg/model"
)

var ErrNotFound = errors.New("movie metadata not found")

type ratingGateway interface {
	GetAggregatedRating(
		ctx context.Context,
		recordID ratingModel.RecordID,
		recordType ratingModel.RecordType,
	) (float64, error)
	PutRating(
		ctx context.Context,
		recordID ratingModel.RecordID,
		recordType ratingModel.RecordType,
		rating *ratingModel.Rating,
	) error
}

type metadataGateway interface {
	Get(ctx context.Context, id string) (*metadataModel.Metadata, error)
}

type Controller struct {
	ratingGateway   ratingGateway
	metadataGateway metadataGateway
}

func New(ratingGateway ratingGateway, metadataGateway metadataGateway) *Controller {
	return &Controller{ratingGateway, metadataGateway}
}

func (c *Controller) Get(ctx context.Context, id string) (*model.MovieDetails, error) {
	metadata, err := c.metadataGateway.Get(ctx, id)
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	details := &model.MovieDetails{Metadata: *metadata}
	rating, err := c.ratingGateway.GetAggregatedRating(ctx, ratingModel.RecordID(id), ratingModel.RecordTypeMovie)
	if err != nil && !errors.Is(err, gateway.ErrNotFound) {
		return nil, err
	} else if err != nil {
		return nil, err
	} else {
		details.Rating = &rating
	}
	return details, nil
}
