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

	"crypto/tls"

	"github.com/JustWorking42/shortener-go-yandex/internal/app"
	accesscontrol "github.com/JustWorking42/shortener-go-yandex/internal/app/accessControl"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/cookie"
	grpcShortener "github.com/JustWorking42/shortener-go-yandex/internal/app/grpc"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/handlers"
	"github.com/JustWorking42/shortener-go-yandex/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
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
	defer app.Repository.CloseDB()

	server := http.Server{
		Addr:    config.ServerAdr,
		Handler: handlers.Webhook(app),
		BaseContext: func(_ net.Listener) context.Context {
			return mainContext
		},
	}

	var grpcServer *grpc.Server
	interceptors := grpc.ChainUnaryInterceptor(
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			if info.FullMethod == "/proto.ShortenerService/GetUserURLs" {
				return cookie.OnlyAuthorizedMiddlewareGRPC(ctx, req, info, handler)
			}
			return handler(ctx, req)
		},
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			if info.FullMethod == "/proto.ShortenerService/GetStats" {
				return accesscontrol.CidrAccessInterceptor(ctx, req, info, handler, app)
			}
			return handler(ctx, req)
		},
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			return cookie.MetadataCheckMiddlewareGRPC(ctx, req, info, handler, app)
		},
	)
	if config.EnableHTTPS {
		certFile := fmt.Sprintf("%s%vcert.pem", config.SSLCertPath, os.PathSeparator)
		keyFile := fmt.Sprintf("%s%vprivate.key", config.SSLCertPath, os.PathSeparator)
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			app.Logger.Sugar().Fatalf("failed to load server key pair: %v", err)
		}
		creds := credentials.NewServerTLSFromCert(&cert)
		grpcServer = grpc.NewServer(grpc.Creds(creds), interceptors)
	} else {
		grpcServer = grpc.NewServer(interceptors)
	}

	proto.RegisterShortenerServiceServer(grpcServer, grpcShortener.NewShortenerService(app))
	reflection.Register(grpcServer)

	wg, errChan := app.Repository.DeleteManager.SubcribeOnTask(mainContext)
	go func() {
		defer close(app.Repository.DeleteManager.TaskChan)
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

	go func() {
		lis, err := net.Listen("tcp", config.GRPCServerAdr)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
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
		grpcServer.GracefulStop()
		app.Logger.Sugar().Info("Server exiting")
	}

}

func printBuildData() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
