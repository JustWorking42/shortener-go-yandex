package sql

import (
	"context"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	db *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, connString string) (*PostgresStorage, error) {
	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Init(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
			short_url TEXT PRIMARY KEY,
			original_url TEXT NOT NULL
		)
	`)
	return err
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *PostgresStorage) Save(ctx context.Context, savedURL storage.SavedURL) error {
	sqlRequest := `INSERT INTO urls (short_url, original_url)
	VALUES ($1, $2)
	ON CONFLICT (short_url) DO NOTHING
`
	_, err := s.db.Exec(ctx, sqlRequest, savedURL.ShortURL, savedURL.OriginalURL)
	return err
}

func (s *PostgresStorage) Get(ctx context.Context, key string) (storage.SavedURL, error) {
	sqlRequest := `SELECT original_url
	FROM urls
	WHERE short_url = $1
`
	row := s.db.QueryRow(ctx, sqlRequest, key)
	var originalURL string
	err := row.Scan(&originalURL)
	if err != nil {
		return storage.SavedURL{}, err
	}

	return storage.SavedURL{ShortURL: key, OriginalURL: originalURL}, nil
}
