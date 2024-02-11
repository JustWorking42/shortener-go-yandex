// Package storage provides an interface for interacting with the application's storage.
package storage

import (
	"context"
	"errors"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
)

// Storage is an interface that defines the methods for interacting with the application's storage.
type Storage interface {

	// Init initializes the storage.
	Init(ctx context.Context) error

	// Ping checks the connection to the storage.
	Ping(ctx context.Context) error

	// Save saves a URL to the storage.
	// It returns the short URL and an error if there was a conflict.
	Save(ctx context.Context, savedURL SavedURL) (string, error)

	// SaveArray saves an array of URLs to the storage.
	SaveArray(ctx context.Context, savedUrls []SavedURL) error

	// Get retrieves a URL from the storage.
	Get(ctx context.Context, key string) (SavedURL, error)

	// IsUserIDExists checks if a user ID exists in the storage.
	IsUserIDExists(ctx context.Context, userID string) (bool, error)

	// GetByUser retrieves all URLs associated with a user.
	GetByUser(ctx context.Context, userID string) ([]SavedURL, error)

	// Delete deletes specified URLs.
	Delete(ctx context.Context, deleteTaskSlice []models.DeleteTask) error

	// Clean cleans the storage.
	Clean(ctx context.Context) error

	// GetStats returns statistics about the storage.
	GetStats(ctx context.Context) (Stats, error)

	// Close closes the storage.
	Close() error
}

// ErrURLConflict is an error that occurs when a URL is already taken.
var ErrURLConflict = errors.New("url is already taken")

// SavedURL represents a saved URL.
type SavedURL struct {
	ShortURL    string `json:"shortUrl"`
	OriginalURL string `json:"originalUrl"`
	UserID      string `json:"userID"`
	IsDeleted   bool
}

type Stats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// NewSavedURL creates a new SavedURL instance.
func NewSavedURL(shortURL string, url string, userID string) *SavedURL {
	return &SavedURL{
		ShortURL:    shortURL,
		OriginalURL: url,
		UserID:      userID,
		IsDeleted:   false,
	}
}
