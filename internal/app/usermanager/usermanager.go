// Package usermanager provides functionalities related to user management.
package usermanager

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
)

// UserManager represents the main user management structure.
type UserManager struct {
	Storage storage.Storage
}

// GenerateUserID generates a unique user ID.
func (um *UserManager) GenerateUserID(ctx context.Context) (string, error) {
	userID, err := generate(ctx, um.Storage)
	if err != nil {
		return "", err
	}

	return userID, nil
}

// generate is a helper function that generates a unique user ID.
// It uses a random number generator to create a unique ID, and checks if the generated ID already exists in the storage.
// If the ID exists, it recursively calls itself until it finds a unique ID.
func generate(ctx context.Context, storage storage.Storage) (string, error) {
	randSlice := make([]byte, 24)
	_, err := rand.Read(randSlice)
	if err != nil {
		return "", err
	}

	userID := base64.StdEncoding.EncodeToString(randSlice)

	exist, err := storage.IsUserIDExists(ctx, userID)
	if err != nil {
		return "", err
	}
	if exist {
		return generate(ctx, storage)
	}

	return userID, nil
}
