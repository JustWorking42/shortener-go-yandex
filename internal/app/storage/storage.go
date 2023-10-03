package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
)

var (
	mutex         sync.Mutex
	serverStorage map[string]string
)

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

func Init() error {
	if configs.GetServerConfig().IsFileStorageEnabled() {
		dir := filepath.Dir(configs.GetServerConfig().FileStoragePath)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
		if !isFileExists(configs.GetServerConfig().FileStoragePath) {
			file, err := os.Create(configs.GetServerConfig().FileStoragePath)
			if err != nil {
				return err
			}
			defer file.Close()
		}
	}
	serverStorage = make(map[string]string)
	return nil
}

func GetStorage() (*map[string]string, error) {
	if serverStorage == nil {
		return nil, errors.New("nil pointer exception call Init before")
	}
	return &serverStorage, nil
}

func Save(model SavedURL) error {
	if !configs.GetServerConfig().IsFileStorageEnabled() {
		mutex.Lock()
		defer mutex.Unlock()
		serverStorage[model.ShortURL] = model.OriginalURL
		return nil
	}
	data, err := json.Marshal(model)
	if err != nil {
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	file, err := os.OpenFile(configs.GetServerConfig().FileStoragePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	if _, err := writer.Write(data); err != nil {
		return err
	}

	if err := writer.WriteByte('\n'); err != nil {
		return err
	}

	return writer.Flush()
}

func Get(key string) (SavedURL, error) {
	var savedURL SavedURL
	if !configs.GetServerConfig().IsFileStorageEnabled() {
		mutex.Lock()
		defer mutex.Unlock()
		link, exist := serverStorage[key]
		if !exist {
			return savedURL, errors.New("link doenst exist")
		}
		savedURL = *NewSavedURL(key, link)
		return savedURL, nil
	}
	mutex.Lock()
	defer mutex.Unlock()
	file, err := os.Open(configs.GetServerConfig().FileStoragePath)
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

func isFileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
