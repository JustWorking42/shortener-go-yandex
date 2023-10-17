package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
)

type MemoryStorage struct {
	store map[string]string
	mu    sync.Mutex
}

func (m *MemoryStorage) Init(ctx context.Context) error {
	if m.store != nil {
		return errors.New("MemoryStorage already initialized")
	}
	m.store = make(map[string]string)
	return nil
}

func (m *MemoryStorage) Ping(ctx context.Context) error {
	if m.store == nil {
		return errors.New("MemoryStorage not initialized")
	}
	return nil
}

func (m *MemoryStorage) Save(ctx context.Context, savedURL storage.SavedURL) (string, error) {
	if m.store == nil {
		return "", errors.New("MemoryStorage not initialized")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for key, item := range m.store {
		if item == savedURL.OriginalURL {
			return key, storage.ErrURLConflict
		}
	}
	m.store[savedURL.ShortURL] = savedURL.OriginalURL
	return "", nil
}

func (m *MemoryStorage) SaveArray(ctx context.Context, savedUrls []storage.SavedURL) error {
	if m.store == nil {
		return errors.New("MemoryStorage not initialized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, url := range savedUrls {
		m.store[url.ShortURL] = url.OriginalURL
	}

	return nil
}

func (m *MemoryStorage) Get(ctx context.Context, key string) (storage.SavedURL, error) {
	if m.store == nil {
		return storage.SavedURL{}, errors.New("MemoryStorage not initialized")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	url, ok := m.store[key]
	if !ok {
		return storage.SavedURL{}, errors.New("URL not found")
	}

	return storage.SavedURL{ShortURL: key, OriginalURL: url}, nil
}

func (m *MemoryStorage) Clean(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k := range m.store {
		delete(m.store, k)
	}
	return nil
}

func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.store = nil
	return nil
}
