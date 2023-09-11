package main

import (
	"net/http"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/handlers"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
)

func main() {
	storage.Init()
	configs.ParseFlags()
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	return http.ListenAndServe(configs.ServerConfig.ServerAdr, handlers.Webhook())
}
