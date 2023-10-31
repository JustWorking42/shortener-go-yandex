package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JustWorking42/shortener-go-yandex/internal/app"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/handlers"
)

func main() {
	mainContext, MainCancel := context.WithCancel(context.Background())
	defer MainCancel()

	config, err := configs.ParseFlags()

	if err != nil {
		log.Fatalf("Parse flags err: %v\n", err)
	}

	app, err := app.CreateApp(mainContext, *config)
	if err != nil {
		log.Fatalf("App init err: %v err", err)
	}
	defer app.Storage.Close()

	server := http.Server{
		Addr:    config.ServerAdr,
		Handler: handlers.Webhook(app),
		BaseContext: func(_ net.Listener) context.Context {
			return mainContext
		},
	}

	app.DeleteManager.SubcribeOnTask(mainContext)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Sugar().Fatalf("Server closed: %v err", err)
		}
	}()

	exit := make(chan os.Signal, 1)

	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit
	app.Logger.Sync()
	app.Logger.Sugar().Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		app.Logger.Sugar().Fatal("Server forced to shutdown:", err)
	}

	app.Logger.Sugar().Info("Server exiting")
}
