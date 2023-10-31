package sql

import (
	"context"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostgresStorage struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewPostgresStorage(ctx context.Context, connString string, logger *zap.Logger) (*PostgresStorage, error) {
	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db, logger: logger}, nil
}

func (s *PostgresStorage) Init(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS urlsTable (
			short_url TEXT PRIMARY KEY,
			original_url TEXT NOT NULL UNIQUE,
			user_id VARCHAR(32) NOT NULL,
			is_deleted bool DEFAULT false
		)
	`)
	if err != nil {
		s.logger.Sugar().Errorf("postgress init error: %v", err)
		return err
	}
	err = s.migration(ctx)
	if err != nil {
		s.logger.Sugar().Errorf("postgress migration error: %v", err)
		return err
	}
	return err
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}
func (s *PostgresStorage) Save(ctx context.Context, savedURL storage.SavedURL) (string, error) {
	sqlRequest := `INSERT INTO urlsTable (short_url, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url RETURNING short_url`
	row := s.db.QueryRow(ctx, sqlRequest, savedURL.ShortURL, savedURL.OriginalURL, savedURL.UserID)
	var shortURL string
	err := row.Scan(&shortURL)
	if err != nil {
		return "", err
	}
	if savedURL.ShortURL != shortURL {
		return shortURL, storage.ErrURLConflict
	}
	s.logger.Sugar().Infof("%v is succesfully save", shortURL)
	return "", nil
}

func (s *PostgresStorage) SaveArray(ctx context.Context, savedUrls []storage.SavedURL) error {
	sqlRequest := `INSERT INTO urlsTable (short_url, original_url, user_id)
	VALUES ($1, $2, $3)
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
		_, err := tx.Exec(ctx, "saveArray", url.ShortURL, url.OriginalURL, url.UserID)
		if err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.logger.Sugar().Infof("%v is succesfully save", savedUrls)
	return nil
}

func (s *PostgresStorage) Get(ctx context.Context, key string) (storage.SavedURL, error) {
	sqlRequest := `SELECT original_url, is_deleted
	FROM urlsTable
	WHERE short_url = $1
`
	row := s.db.QueryRow(ctx, sqlRequest, key)
	var originalURL string
	var isDeleted bool
	err := row.Scan(&originalURL, &isDeleted)
	if err != nil {
		s.logger.Sugar().Errorf("postgress get error: %v", err)
		return storage.SavedURL{}, err
	}

	return storage.SavedURL{ShortURL: key, OriginalURL: originalURL, IsDeleted: isDeleted}, nil
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

func (s *PostgresStorage) IsUserIDExists(ctx context.Context, userID string) (bool, error) {
	sqlRequest := `SELECT EXISTS(SELECT 1 FROM urlsTable WHERE user_id=$1)`
	row := s.db.QueryRow(ctx, sqlRequest, userID)
	var exists bool
	err := row.Scan(&exists)
	s.logger.Sugar().Infof("postgress is user exitst: %v", exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *PostgresStorage) GetByUser(ctx context.Context, userID string) ([]storage.SavedURL, error) {
	rows, err := s.db.Query(ctx, "SELECT * FROM urlsTable WHERE user_id=$1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var savedURLs []storage.SavedURL
	for rows.Next() {
		var savedURL storage.SavedURL
		err = rows.Scan(&savedURL.ShortURL, &savedURL.OriginalURL, &savedURL.UserID, &savedURL.IsDeleted)
		if err != nil {
			return nil, err
		}
		savedURLs = append(savedURLs, savedURL)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return savedURLs, nil
}

func (s *PostgresStorage) Delete(ctx context.Context, taskSlice []models.DeleteTask) error {
	b := &pgx.Batch{}
	for _, task := range taskSlice {
		b.Queue(`UPDATE urlsTable SET is_deleted = true WHERE short_url = $1 AND user_id = $2`, task.URL, task.UserID)
	}

	br := s.db.SendBatch(ctx, b)
	defer br.Close()

	for i := 0; i < len(taskSlice); i++ {
		_, err := br.Exec()
		if err != nil {
			s.logger.Sugar().Errorf("error while deliting %v", err)
			return err
		}
	}

	return nil
}

func (s *PostgresStorage) migration(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		ALTER TABLE urlsTable
		ADD COLUMN IF NOT EXISTS user_id VARCHAR(32) NOT NULL
	`)
	if err != nil {
		return err

	}
	return s.secondMigration(ctx)
}

func (s *PostgresStorage) secondMigration(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
	ALTER TABLE urlsTable
	ADD COLUMN IF NOT EXISTS is_deleted bool DEFAULT false
`)
	return err
}
