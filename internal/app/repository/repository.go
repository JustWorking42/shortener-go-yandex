// Package repository provides an interface for interacting with the storage layer
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/deletemanager"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/urlgenerator"
)

// Repository represents the data access layer of the application.
type Repository struct {
	storage       storage.Storage
	DeleteManager *deletemanager.DeleteManager
	redirectHost  string
}

// NewRepository creates a new instance of the Repository with the given storage.
func NewRepository(storage storage.Storage, deletemanager *deletemanager.DeleteManager, redirectHost string) *Repository {
	return &Repository{
		storage:       storage,
		DeleteManager: deletemanager,
		redirectHost:  redirectHost,
	}
}

// SaveURL saves a URL to the storage and returns the saved URL.
func (r *Repository) SaveURL(ctx context.Context, originalURL, userID string) (storage.SavedURL, error) {
	shortID := urlgenerator.CreateShortLink()
	savedURL := storage.NewSavedURL(shortID, originalURL, userID)

	conflictURL, err := r.storage.Save(ctx, *savedURL)
	if err != nil {
		if errors.Is(err, storage.ErrURLConflict) {
			return storage.SavedURL{ShortURL: conflictURL}, err
		}
		return storage.SavedURL{}, err
	}

	return storage.SavedURL{ShortURL: shortID}, nil
}

// DeleteURLs deletes the specified URLs.
func (r *Repository) DeleteURLs(ctx context.Context, userID string, urls []string) error {
	for _, url := range urls {
		r.DeleteManager.TaskChan <- models.DeleteTask{
			UserID: userID,
			URL:    url,
		}
	}
	return nil
}

// GetUserURLs retrieves all URLs associated with a user.
func (r *Repository) GetUserURLs(ctx context.Context, userID string) ([]storage.SavedURL, error) {
	urls, err := r.storage.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return urls, nil
}

// GetURL retrieves a URL by its short ID.
func (r *Repository) GetURL(ctx context.Context, id string) (storage.SavedURL, error) {
	savedURL, err := r.storage.Get(ctx, id)
	if err != nil {
		return storage.SavedURL{}, err
	}
	return savedURL, nil
}

// GetStats retrieves statistics for URLs and users.
func (r *Repository) GetStats(ctx context.Context) (storage.Stats, error) {
	stats, err := r.storage.GetStats(ctx)
	if err != nil {
		return storage.Stats{}, err
	}
	return stats, nil
}

// SaveURLArray saves an array of URLs to the storage and returns the saved URLs.
func (r *Repository) SaveURLArray(ctx context.Context, urls []models.RequestShortenerURLBatch, userID string) ([]models.ResponseShortenerURLBatch, error) {
	var savedURLs []models.ResponseShortenerURLBatch
	var savedURLsData []storage.SavedURL

	for _, item := range urls {
		if item.URL == "" {
			return nil, errors.New("original url is empty check request body")
		}
		shortID := urlgenerator.CreateShortLink()
		savedURL := storage.NewSavedURL(shortID, item.URL, userID)
		savedURLsData = append(savedURLsData, *savedURL)
		savedURLs = append(savedURLs, *models.NewResponseShortenerURLBatch(item.ID, fmt.Sprintf("%s/%s", r.redirectHost, shortID)))
	}

	err := r.storage.SaveArray(ctx, savedURLsData)
	if err != nil {
		return nil, err
	}

	return savedURLs, nil
}

// PingDB checks the connectivity to the database by pinging it.
func (r *Repository) PingDB(ctx context.Context) error {
	return r.storage.Ping(ctx)
}

// CloseDB closes the connection to the database.
func (r *Repository) CloseDB() error {
	return r.storage.Close()
}
