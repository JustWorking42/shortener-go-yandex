// Package memory provides functionality for storing and retrieving URLs in memory.
package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
)

// MemoryStorage represents a memory storage for URLs.
type MemoryStorage struct {
	store []storage.SavedURL
	mu    sync.Mutex
}

// Init initializes the memory storage.
func (m *MemoryStorage) Init(ctx context.Context) error {
	if m.store != nil {
		return errors.New("MemoryStorage already initialized")
	}
	m.store = []storage.SavedURL{}
	return nil
}

// Ping checks if the memory storage is initialized.
func (m *MemoryStorage) Ping(ctx context.Context) error {
	if m.store == nil {
		return errors.New("MemoryStorage not initialized")
	}
	return nil
}

// Delete deletes a URL from the memory storage.
func (m *MemoryStorage) Delete(ctx context.Context, taskSlice []models.DeleteTask) error {

	for i, item := range m.store {
		for _, task := range taskSlice {
			if item.ShortURL == task.URL && item.UserID == task.UserID {
				m.store[i].IsDeleted = true
			}
		}
	}
	return nil
}

// Save saves a URL to the memory storage.
// It returns the short URL and an error if there was a conflict.
func (m *MemoryStorage) Save(ctx context.Context, savedURL storage.SavedURL) (string, error) {
	if m.store == nil {
		return "", errors.New("MemoryStorage not initialized")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, item := range m.store {
		if item.OriginalURL == savedURL.OriginalURL {
			return item.ShortURL, storage.ErrURLConflict
		}
	}
	m.store = append(m.store, savedURL)
	return "", nil
}

// SaveArray saves an array of URLs to the memory storage.
func (m *MemoryStorage) SaveArray(ctx context.Context, savedUrls []storage.SavedURL) error {
	if m.store == nil {
		return errors.New("MemoryStorage not initialized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = append(m.store, savedUrls...)

	return nil
}

// Get gets a URL from the memory storage by its short URL.
func (m *MemoryStorage) Get(ctx context.Context, key string) (storage.SavedURL, error) {
	if m.store == nil {
		return storage.SavedURL{}, errors.New("MemoryStorage not initialized")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, item := range m.store {
		if item.ShortURL == key {
			return item, nil
		}
	}

	return storage.SavedURL{}, errors.New("URL not found")
}

// Clean cleans the memory storage.
func (m *MemoryStorage) Clean(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = []storage.SavedURL{}
	return nil
}

// IsUserIDExists checks if a user ID exists in the memory storage.
func (m *MemoryStorage) IsUserIDExists(ctx context.Context, userID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.store == nil {
		return false, errors.New("MemoryStorage not initialized")
	}
	for _, item := range m.store {
		if item.UserID == userID {
			return true, nil
		}
	}

	return false, nil
}

// Close closes the memory storage.
func (m *MemoryStorage) Close() error {
	return nil
}

// GetByUser gets all URLs associated with a user ID from the memory storage.
func (m *MemoryStorage) GetByUser(ctx context.Context, userID string) ([]storage.SavedURL, error) {
	if m.store == nil {
		return nil, errors.New("MemoryStorage not initialized")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var userURLs []storage.SavedURL
	for _, item := range m.store {
		if item.UserID == userID {
			userURLs = append(userURLs, item)
		}
	}

	if len(userURLs) == 0 {
		return nil, errors.New("no URLs found for this user")
	}

	return userURLs, nil
}

// GetStats returns returns the number of users and urls in the memory storage.
func (m *MemoryStorage) GetStats(ctx context.Context) (storage.Stats, error) {
	if m.store == nil {
		return storage.Stats{}, errors.New("MemoryStorage not initialized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	var usersMap = make(map[string]bool)
	var urls int

	for _, item := range m.store {
		if !item.IsDeleted {
			urls++
			usersMap[item.UserID] = true
		}
	}
	var users int

	for _ = range usersMap {
		users++
	}

	return storage.Stats{
		URLs:  urls,
		Users: users,
	}, nil
}
