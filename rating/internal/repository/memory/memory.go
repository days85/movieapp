package memory

import (
	"context"

	"movieexample.com/rating/internal/repository"
	"movieexample.com/rating/pkg/model"
)

type Repository struct {
	data map[model.RecordType]map[model.RecordID][]model.Rating
}

func New() *Repository {
	return &Repository{map[model.RecordType]map[model.RecordID][]model.Rating{}}
}

func (r *Repository) Get(
	ctx context.Context,
	recordID model.RecordID,
	recordType model.RecordType,
) ([]model.Rating, error) {
	if _, ok := r.data[recordType]; !ok {
		return nil, repository.ErrNotFound
	}
	return r.data[recordType][recordID], nil
}

func (r *Repository) Put(
	ctx context.Context,
	recordID model.RecordID,
	recordType model.RecordType,
	rating *model.Rating,
) error {
	if _, ok := r.data[recordType]; !ok {
		r.data[recordType] = map[model.RecordID][]model.Rating{}
	}
	r.data[recordType][recordID] = append(r.data[recordType][recordID], *rating)
	return nil
}
