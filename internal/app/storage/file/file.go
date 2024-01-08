// Package file provides functionality for storing and retrieving URLs in a file.
package file

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
)

// FileStorage represents a file storage for URLs.
type FileStorage struct {
	FilePath string
	File     *os.File
	mu       sync.Mutex
}

// Init initializes the file storage.
// It creates the directory if it doesn't exist and opens the file.
func (fs *FileStorage) Init(ctx context.Context) error {
	dir := filepath.Dir(fs.FilePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	if !fs.isFileExists() {
		fs.File, err = os.Create(fs.FilePath)
		if err != nil {
			return err
		}
	} else {
		fs.File, err = os.OpenFile(fs.FilePath, os.O_RDWR, 0666)
		if err != nil {
			return err
		}
	}
	return nil
}

// isFileExists checks if the file exists.
func (fs *FileStorage) isFileExists() bool {
	_, err := os.Stat(fs.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// Ping checks if the file is open.
func (fs *FileStorage) Ping(ctx context.Context) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if !fs.isFileExists() {
		return errors.New("file does not exist")
	}

	if fs.File == nil {
		return errors.New("file does not open")
	}

	return nil
}

// Save saves a URL to the file.
// It returns the short URL and an error if there was a conflict.
func (fs *FileStorage) Save(ctx context.Context, savedURL storage.SavedURL) (string, error) {

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.File == nil {
		return "", errors.New("file does not open")
	}

	scanner := bufio.NewScanner(fs.File)
	for scanner.Scan() {
		var url storage.SavedURL
		if strings.Contains(scanner.Text(), savedURL.OriginalURL) {
			err := json.Unmarshal(scanner.Bytes(), &url)
			if err != nil {
				return "", err
			}
			return url.ShortURL, storage.ErrURLConflict
		}
	}

	data, err := json.Marshal(savedURL)
	if err != nil {
		return "", err
	}

	writer := bufio.NewWriter(fs.File)
	if _, err := writer.Write(data); err != nil {
		return "", err
	}

	if err := writer.WriteByte('\n'); err != nil {
		return "", err
	}

	return "", writer.Flush()
}

// SaveArray saves an array of URLs to the file.
func (fs *FileStorage) SaveArray(ctx context.Context, savedUrls []storage.SavedURL) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.File == nil {
		return errors.New("file does not open")
	}

	writer := bufio.NewWriter(fs.File)

	for _, url := range savedUrls {
		data, err := json.Marshal(url)
		if err != nil {
			return err
		}

		if _, err := writer.Write(data); err != nil {
			return err
		}

		if err := writer.WriteByte('\n'); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// Get gets a URL from the file by its short URL.
func (fs *FileStorage) Get(ctx context.Context, key string) (storage.SavedURL, error) {
	var savedURL storage.SavedURL
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.File == nil {
		return savedURL, errors.New("file does not open")
	}

	_, err := fs.File.Seek(0, 0)
	if err != nil {
		return savedURL, err
	}

	scanner := bufio.NewScanner(fs.File)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), key) {
			json.Unmarshal(scanner.Bytes(), &savedURL)
			return savedURL, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return savedURL, err
	}

	return savedURL, errors.New("Nothing")
}

// Delete deletes a URL from the file.
func (fs *FileStorage) Delete(ctx context.Context, taskSlice []models.DeleteTask) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.File == nil {
		return errors.New("file does not open")
	}

	_, err := fs.File.Seek(0, 0)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(fs.File)

	var buffer []storage.SavedURL

	for scanner.Scan() {
		for _, task := range taskSlice {
			var savedURL storage.SavedURL
			json.Unmarshal(scanner.Bytes(), &savedURL)
			if savedURL.ShortURL == task.URL && savedURL.UserID == task.UserID {
				savedURL.IsDeleted = true
			}
			buffer = append(buffer, savedURL)

		}
	}
	if err := fs.Clean(ctx); err != nil {
		return err
	}
	return fs.SaveArray(ctx, buffer)
}

// Clean cleans the file.
func (fs *FileStorage) Clean(ctx context.Context) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.File == nil {
		return errors.New("file does not open")
	}

	if err := fs.File.Truncate(0); err != nil {
		return err
	}

	if _, err := fs.File.Seek(0, 0); err != nil {
		return err
	}
	return nil
}

// IsUserIDExists checks if a user ID exists in the file.
func (fs *FileStorage) IsUserIDExists(ctx context.Context, userID string) (bool, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.File == nil {
		return false, errors.New("file does not open")
	}

	_, err := fs.File.Seek(0, 0)
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(fs.File)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), userID) {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// GetByUser gets all URLs associated with a user ID from the file.
func (fs *FileStorage) GetByUser(ctx context.Context, userID string) ([]storage.SavedURL, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.File == nil {
		return nil, errors.New("file does not open")
	}

	_, err := fs.File.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fs.File)
	var urls []storage.SavedURL
	var savedURL storage.SavedURL
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), userID) {
			err := json.Unmarshal(scanner.Bytes(), &savedURL)
			if err != nil {
				return nil, err
			}
			urls = append(urls, savedURL)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(urls) == 0 {
		return nil, errors.New("no URLs found for this user")
	}

	return urls, nil
}

// Close closes the file.
func (fs *FileStorage) Close() error {
	if fs.File != nil {
		return fs.File.Close()
	}
	return nil
}
