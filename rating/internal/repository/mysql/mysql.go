package mysql

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"movieexample.com/rating/internal/repository"
	"movieexample.com/rating/pkg/model"
)

type Repository struct {
	db *sql.DB
}

func New() (*Repository, error) {
	db, err := sql.Open("mysql", "root:password@tcp(movieapp_db:3306)/movieexample")
	if err != nil {
		return nil, err
	}
	return &Repository{db}, nil
}

func (r *Repository) Get(
	ctx context.Context,
	recodID model.RecordID,
	recordType model.RecordType,
) ([]model.Rating, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT user_id, value FROM ratings WHERE record_id = ? AND record_type = ?",
		recodID,
		recordType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []model.Rating
	for rows.Next() {
		var userID string
		var value int32
		if err := rows.Scan(&userID, &value); err != nil {
			return nil, err
		}
		res = append(res, model.Rating{
			UserID: model.UserID(userID),
			Value:  model.RatingValue(value),
		})
	}
	if len(res) == 0 {
		return nil, repository.ErrNotFound
	}
	return res, nil
}

func (r *Repository) Put(
	ctx context.Context,
	recordID model.RecordID,
	recordType model.RecordType,
	rating *model.Rating,
) error {
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO ratings (record_id, record_type, user_id, value) VALUES (?, ?, ?, ?)",
		recordID,
		recordType,
		rating.UserID,
		rating.Value,
	)
	return err
}
