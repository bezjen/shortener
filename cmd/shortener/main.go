package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/config/db"
	grpchandler "github.com/bezjen/shortener/internal/grpc"
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/bezjen/shortener/internal/router"
	"github.com/bezjen/shortener/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Global build information variables
// These are set during build process using ldflags
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

const grpcPort = ":3200"

func main() {
	printBuildInfo()
	config.ParseConfig()
	cfg := config.AppConfig

	shortenerLogger, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Error during logger initialization: %v", err)
	}

	storage, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Error during storage initialization: %v", err)
	}
	defer func(storage repository.Repository) {
		err = storage.Close()
		if err != nil {
			log.Printf("Error during storage close cleanly: %v", err)
		}
	}(storage)
	urlShortener := service.NewURLShortener(storage, shortenerLogger)
	defer urlShortener.Close()
	authorizer := service.NewAuthorizer([]byte(cfg.SecretKey), shortenerLogger)
	auditService := service.NewShortenerAuditService(shortenerLogger)
	auditService.ConfigureObservers(cfg)
	shortenerHandler := handler.NewShortenerHandler(cfg, shortenerLogger, urlShortener, auditService)
	shortenerRouter := router.NewRouter(shortenerLogger, authorizer, *shortenerHandler)
	httpServer := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: shortenerRouter,
	}

	grpcHandler := grpchandler.NewHandler(cfg, shortenerLogger, authorizer, urlShortener, auditService)
	listen, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcHandler.UnaryInterceptorMiddleware()),
	)
	grpcHandler.RegisterService(grpcServer)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		var err error
		if cfg.EnableHTTPS {
			err = httpServer.ListenAndServeTLS("./server.crt", "./server.key")
		} else {
			err = httpServer.ListenAndServe()
		}
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP Server failed: %v", err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		shortenerLogger.Info("Starting gRPC server", zap.String("addr", grpcPort))

		if err := grpcServer.Serve(listen); err != nil {
			log.Printf("gRPC Server failed: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP Server forced to shutdown: %v", err)
	}
	grpcServer.GracefulStop()

	wg.Wait()
}

// printBuildInfo outputs build version, date and commit information
func printBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}

	date := buildDate
	if date == "" {
		date = "N/A"
	}

	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	log.Printf("Build version: %s\n", version)
	log.Printf("Build date: %s\n", date)
	log.Printf("Build commit: %s\n", commit)
}
