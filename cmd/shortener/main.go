package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JustWorking42/shortener-go-yandex/internal/app"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/handlers"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/logger"
)

func main() {

	config, err := configs.ParseFlags()
	if err != nil {
		logger.Log.Sugar().Fatalf("Parse flags err: %v err", err)
	}

	app, err := app.CreateApp(context.Background(), *config)
	if err != nil {
		logger.Log.Sugar().Fatalf("Server storage init err: %v err", err)
	}

	if err := logger.Init(config.LogLevel); err != nil {
		logger.Log.Sugar().Fatalf("Init logger err: %v err", err)
	}

	server := http.Server{
		Addr:    config.ServerAdr,
		Handler: handlers.Webhook(app),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Sugar().Fatalf("Server closed: %v err", err)
		}
	}()

	exit := make(chan os.Signal, 1)

	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit

	logger.Log.Sugar().Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Sugar().Fatal("Server forced to shutdown:", err)
	}

	logger.Log.Sugar().Info("Server exiting")
}
