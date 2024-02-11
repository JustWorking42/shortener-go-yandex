// Package app provides the main application structure and related functionalities.
package app

import (
	"context"
	"fmt"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/deletemanager"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/logger"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage/file"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage/memory"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage/sql"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/usermanager"
	"go.uber.org/zap"
)

// App represents the main application structure.
// It includes fields for storage, logger, context, user manager, delete manager, and redirect host.
type App struct {
	Storage       storage.Storage
	Logger        *zap.Logger
	context       context.Context
	UserManager   *usermanager.UserManager
	DeleteManager *deletemanager.DeleteManager
	RedirectHost  string
	TrustedSubnet string
}

// CreateApp creates a new instance of the App object.
func CreateApp(ctx context.Context, conf configs.Config) (*App, error) {

	logger, err := logger.CreateLogger(conf.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error creating App %v", err)
	}

	storage, err := createStorage(ctx, conf, logger)

	if err != nil {
		return nil, fmt.Errorf("error creating App %v", err)
	}

	usermanager := &usermanager.UserManager{Storage: storage}

	deletemanager := deletemanager.NewDeleteManager(storage)

	return &App{
		Storage:       storage,
		Logger:        logger,
		context:       ctx,
		UserManager:   usermanager,
		DeleteManager: deletemanager,
		RedirectHost:  conf.RedirectHost,
		TrustedSubnet: conf.TrustedSubnet,
	}, nil
}

// createStorage initializes the storage based on the configuration.
func createStorage(ctx context.Context, conf configs.Config, logger *zap.Logger) (storage.Storage, error) {
	var storage storage.Storage
	var err error

	if conf.DBAddress != "" {
		storage, err = sql.NewPostgresStorage(ctx, conf.DBAddress, logger)
		if err != nil {
			return nil, err
		}
	} else if path := conf.FileStoragePath; path != "" {
		storage = &file.FileStorage{FilePath: path}
	} else {
		storage = &memory.MemoryStorage{}
	}
	storage.Init(ctx)
	return storage, nil
}
