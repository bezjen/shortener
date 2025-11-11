package main

import (
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/config/db"
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/bezjen/shortener/internal/router"
	"github.com/bezjen/shortener/internal/service"
	"log"
	"net/http"
	_ "net/http/pprof"
)

// Global build information variables
// These are set during build process using ldflags
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err = http.ListenAndServe(cfg.ServerAddr, shortenerRouter); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server failed to start: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
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
