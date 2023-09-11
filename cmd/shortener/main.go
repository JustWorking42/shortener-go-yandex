package main

import (
	"log"
	"net/http"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/handlers"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
)

func main() {
	if err := storage.Init(); err != nil {
		log.Fatalf("Server storage init err: %v err", err)
	}

	if err := configs.ParseFlags(); err != nil {
		log.Fatalf("Parse flags err: %v err", err)
	}

	if err := run(); err != nil {
		log.Fatalf("Server closed: %v err", err)
	}
}

func run() error {
	return http.ListenAndServe(configs.GetServerConfig().ServerAdr, handlers.Webhook())
}
