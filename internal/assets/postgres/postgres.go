package postgres

import (
	"github.com/alesr/platform-go-challenge/internal/assets"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ assets.Repository = (*Repository)(nil)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}
