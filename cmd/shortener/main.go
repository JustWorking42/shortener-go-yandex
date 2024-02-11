package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/JustWorking42/shortener-go-yandex/internal/app"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/handlers"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	printBuildData()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	mainContext, MainCancel := context.WithCancel(context.Background())
	defer MainCancel()

	config, err := configs.ParseFlags()

	if err != nil {
		if errors.Is(err, configs.ErrParseConfigJson) {
			log.Println(err.Error())
		} else {
			log.Fatalf("Parse flags err: %v\n", err)
		}
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

	wg, errChan := app.DeleteManager.SubcribeOnTask(mainContext)
	go func() {
		defer close(app.DeleteManager.TaskChan)
		wg.Wait()
	}()

	go func() {
		var err error
		if config.EnableHTTPS {
			certFile := fmt.Sprintf("%s%vcert.pem", config.SSLCertPath, os.PathSeparator)
			keyFile := fmt.Sprintf("%s%vprivate.key", config.SSLCertPath, os.PathSeparator)
			err = server.ListenAndServeTLS(certFile, keyFile)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Sugar().Fatalf("Server closed: %v err", err)
		}
	}()

	exit := make(chan os.Signal, 1)

	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		app.Logger.Sugar().Fatalf("Delete url err: %v", err)

	case <-exit:
		app.Logger.Sync()
		app.Logger.Sugar().Info("Shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			app.Logger.Sugar().Fatal("Server forced to shutdown:", err)
		}
		app.Logger.Sugar().Info("Server exiting")
	}

}

func printBuildData() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
