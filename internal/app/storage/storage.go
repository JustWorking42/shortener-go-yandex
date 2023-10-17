package storage

import (
	"context"
	"errors"
)

type Storage interface {
	Init(ctx context.Context) error
	Ping(ctx context.Context) error
	Save(ctx context.Context, savedURL SavedURL) (string, error)
	SaveArray(ctx context.Context, savedUrls []SavedURL) error
	Get(ctx context.Context, key string) (SavedURL, error)
	Clean(ctx context.Context) error
	Close() error
}

var ErrURLConflict = errors.New("url is already taken")

type SavedURL struct {
	ShortURL    string `json:"shortUrl"`
	OriginalURL string `json:"originalUrl"`
}

func NewSavedURL(shortURL string, url string) *SavedURL {
	return &SavedURL{
		ShortURL:    shortURL,
		OriginalURL: url,
	}
}
