package sql

import (
	"context"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresStorage struct {
	db *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, connString string) (*PostgresStorage, error) {
	db, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Init(ctx context.Context) error {
	return nil
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *PostgresStorage) Save(ctx context.Context, savedURL storage.SavedURL) error {
	return nil
}

func (s *PostgresStorage) Get(ctx context.Context, key string) (storage.SavedURL, error) {
	return storage.SavedURL{}, nil
}
