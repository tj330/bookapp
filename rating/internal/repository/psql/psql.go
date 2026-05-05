package psql

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tj330/bookapp/rating/pkg/model"
)

type Repository struct {
	db *sql.DB
}

func New() (*Repository, error) {
	connStr := "user=admin password=secret host=localhost port=5432 dbname=book sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT user_id, value FROM ratings WHERE record_id = $1 AND record_type = $2", recordID, recordType)
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
	return res, nil
}

func (r *Repository) Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	if rating == nil {
		return errors.New("rating is nil")
	}

	// Step 1: Try UPDATE
	res, err := r.db.ExecContext(ctx, `
		UPDATE ratings
		SET value = $1
		WHERE record_id = $2 AND user_id = $3
	`, rating.Value, recordID, rating.UserID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	// Step 2: If no rows updated → INSERT
	if rowsAffected == 0 {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO ratings (record_id, record_type, user_id, value)
			VALUES ($1, $2, $3, $4)
		`, recordID, recordType, rating.UserID, rating.Value)
		if err != nil {
			return err
		}
	}

	return nil
}
