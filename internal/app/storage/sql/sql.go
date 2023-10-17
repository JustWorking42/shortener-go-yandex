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
		CREATE TABLE IF NOT EXISTS urlsTable (
			short_url TEXT PRIMARY KEY,
			original_url TEXT NOT NULL UNIQUE
		)
	`)
	return err
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *PostgresStorage) Save(ctx context.Context, savedURL storage.SavedURL) (string, error) {

	sqlRequest := `INSERT INTO urlsTable (short_url, original_url) VALUES ($1, $2) ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url RETURNING short_url`
	row := s.db.QueryRow(ctx, sqlRequest, savedURL.ShortURL, savedURL.OriginalURL)
	var shortURL string
	err := row.Scan(&shortURL)
	if err != nil {
		return "", err
	}
	if savedURL.ShortURL != shortURL {
		return shortURL, storage.ErrURLConflict
	}

	return "", nil
}

func (s *PostgresStorage) SaveArray(ctx context.Context, savedUrls []storage.SavedURL) error {
	sqlRequest := `INSERT INTO urlsTable (short_url, original_url)
	VALUES ($1, $2)
	ON CONFLICT (short_url) DO NOTHING`
	tx, err := s.db.Begin(ctx)

	defer tx.Rollback(ctx)
	if err != nil {
		return err
	}
	_, err = tx.Prepare(ctx, "saveArray", sqlRequest)

	if err != nil {
		return err
	}
	for _, url := range savedUrls {
		_, err := tx.Exec(ctx, "saveArray", url.ShortURL, url.OriginalURL)
		if err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) Get(ctx context.Context, key string) (storage.SavedURL, error) {
	sqlRequest := `SELECT original_url
	FROM urlsTable
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

func (s *PostgresStorage) Clean(ctx context.Context) error {
	sqlRequest := `TRUNCATE TABLE urlsTable`
	_, err := s.db.Exec(ctx, sqlRequest)
	return err
}

func (s *PostgresStorage) Close() error {
	if s.db != nil {
		s.db.Close()
	}
	return nil
}
