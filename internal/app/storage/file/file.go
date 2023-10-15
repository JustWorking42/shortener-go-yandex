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

	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
)

type FileStorage struct {
	FilePath string
	mu       sync.RWMutex
}

func (fs *FileStorage) Init(ctx context.Context) error {
	dir := filepath.Dir(fs.FilePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	if !fs.isFileExists() {
		file, err := os.Create(fs.FilePath)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

func (fs *FileStorage) isFileExists() bool {
	_, err := os.Stat(fs.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	if !fs.isFileExists() {
		return errors.New("file does not exist")
	}

	return nil
}

func (fs *FileStorage) Save(ctx context.Context, model storage.SavedURL) error {
	data, err := json.Marshal(model)
	if err != nil {
		return err
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()
	file, err := os.OpenFile(fs.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	if _, err := writer.Write(data); err != nil {
		return err
	}

	if err := writer.WriteByte('\n'); err != nil {
		return err
	}

	return writer.Flush()
}

func (fs *FileStorage) Get(ctx context.Context, key string) (storage.SavedURL, error) {
	var savedURL storage.SavedURL
	fs.mu.Lock()
	defer fs.mu.Unlock()
	file, err := os.Open(fs.FilePath)
	if err != nil {
		return savedURL, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

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
