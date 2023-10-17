package app

import (
	"context"
	"fmt"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/logger"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage/file"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage/memory"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage/sql"
	"go.uber.org/zap"
)

type App struct {
	Storage      storage.Storage
	Logger       *zap.Logger
	context      context.Context
	RedirectHost string
}

func CreateApp(ctx context.Context, conf configs.Config) (*App, error) {
	storage, err := createStorage(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("error creating App %v", err)
	}

	logger, err := logger.CreateLogger(conf.LogLevel)

	if err != nil {
		return nil, fmt.Errorf("error creating App %v", err)
	}
	return &App{
		Storage:      storage,
		Logger:       logger,
		context:      ctx,
		RedirectHost: conf.RedirectHost,
	}, nil
}

func createStorage(ctx context.Context, conf configs.Config) (storage.Storage, error) {
	var storage storage.Storage
	var err error

	if conf.DBAddress != "" {
		storage, err = sql.NewPostgresStorage(ctx, conf.DBAddress)
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
