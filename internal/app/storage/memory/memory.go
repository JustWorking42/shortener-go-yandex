package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
)

type MemoryStorage struct {
	store []storage.SavedURL
	mu    sync.Mutex
}

func (m *MemoryStorage) Init(ctx context.Context) error {
	if m.store != nil {
		return errors.New("MemoryStorage already initialized")
	}
	m.store = []storage.SavedURL{}
	return nil
}

func (m *MemoryStorage) Ping(ctx context.Context) error {
	if m.store == nil {
		return errors.New("MemoryStorage not initialized")
	}
	return nil
}

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

func (m *MemoryStorage) SaveArray(ctx context.Context, savedUrls []storage.SavedURL) error {
	if m.store == nil {
		return errors.New("MemoryStorage not initialized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = append(m.store, savedUrls...)

	return nil
}

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

func (m *MemoryStorage) Clean(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = []storage.SavedURL{}
	return nil
}

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

func (m *MemoryStorage) Close() error {
	return nil
}

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
