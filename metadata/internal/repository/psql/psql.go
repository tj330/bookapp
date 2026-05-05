package psql

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tj330/bookapp/metadata/internal/repository"
	"github.com/tj330/bookapp/metadata/pkg/model"
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

func (r *Repository) Get(ctx context.Context, id string) (*model.Metadata, error) {
	var title, description, author, isbn string
	row := r.db.QueryRowContext(ctx, "SELECT title, description, author, isbn from books where id=$1", id)
	if err := row.Scan(&title, &description, &author, &isbn); err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &model.Metadata{
		ID:          id,
		Title:       title,
		Description: description,
		Author:      author,
		ISBN:        isbn,
	}, nil
}

func (r *Repository) Put(ctx context.Context, id string, metadata *model.Metadata) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO books (id, title, description, author, isbn) VALUES ($1, $2, $3, $4, $5)",
		id, metadata.Title, metadata.Description, metadata.Author, metadata.ISBN,
	)
	return err
}
