package usermanager

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
)

type UserManager struct {
	Storage storage.Storage
}

func (um *UserManager) GenerateUserID(ctx context.Context) (string, error) {
	userID, err := generate(ctx, um.Storage)
	if err != nil {
		return "", err
	}

	return string(userID), nil
}

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
