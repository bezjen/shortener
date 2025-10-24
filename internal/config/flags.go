// Package config provides configuration management for the URL shortening service.
// It handles command-line flags, environment variables, and configuration defaults.
package config

import (
	"flag"
	"os"
)

// Config holds all application configuration settings.
// Settings can be provided via command-line flags or environment variables.
type Config struct {
	ServerAddr      string // Server address in format "host:port"
	BaseURL         string // Base URL for shortened links
	LogLevel        string // Logging level (debug, info, warn, error)
	FileStoragePath string // Path to file storage (if using file backend)
	DatabaseDSN     string // PostgreSQL connection string
	SecretKey       string // JWT secret key for authentication
	AuditFile       string // Path to audit log file
	AuditURL        string // URL for remote audit logging
}

// AppConfig is the global application configuration instance.
var AppConfig Config

// ParseConfig parses command-line flags and environment variables to populate AppConfig.
// Environment variables take precedence over command-line flags.
// Default values are provided for all configuration options.
//
// Supported environment variables:
//   - SERVER_ADDRESS: server address (overrides -a flag)
//   - BASE_URL: base URL for shortened links (overrides -b flag)
//   - LOG_LEVEL: logging level (overrides -l flag)
//   - FILE_STORAGE_PATH: file storage path (overrides -f flag)
//   - DATABASE_DSN: PostgreSQL connection string (overrides -d flag)
//   - SECRET_KEY: JWT secret key (overrides -s flag)
//   - AUDIT_FILE: audit file path (overrides -audit-file flag)
//   - AUDIT_URL: audit service URL (overrides -audit-url flag)
//
// Command-line flags:
//   - -a: server address (default: "localhost:8080")
//   - -b: base URL (default: "http://localhost:8080")
//   - -l: log level (default: "info")
//   - -f: file storage path (default: "")
//   - -d: database DSN (default: "")
//   - -s: secret key (default: "")
//   - -audit-file: audit file path (default: "")
//   - -audit-url: audit service URL (default: "")
func ParseConfig() {
	flagServerAddr := flag.String("a", "localhost:8080", "port to run server")
	flagBaseURL := flag.String("b", "http://localhost:8080", "address and port of tiny url")
	flagLogLevel := flag.String("l", "info", "log level")
	flagFileStoragePath := flag.String("f", "", "path to file with data")
	flagDatabaseDSN := flag.String("d", "", "postgres data source name in format `postgres://username:password@host:port/database_name?sslmode=disable`")
	flagSecretKey := flag.String("s", "", "authorization secret key")
	flagAuditFile := flag.String("audit-file", "", "path to audit file")
	flagAuditURL := flag.String("audit-url", "", "audit url")
	flag.Parse()

	addr, addrExists := os.LookupEnv("SERVER_ADDRESS")
	if addrExists {
		AppConfig.ServerAddr = addr
	} else {
		AppConfig.ServerAddr = *flagServerAddr
	}
	baseURL, baseURLExists := os.LookupEnv("BASE_URL")
	if baseURLExists {
		AppConfig.BaseURL = baseURL
	} else {
		AppConfig.BaseURL = *flagBaseURL
	}
	logLevel, logLevelExists := os.LookupEnv("LOG_LEVEL")
	if logLevelExists {
		AppConfig.LogLevel = logLevel
	} else {
		AppConfig.LogLevel = *flagLogLevel
	}
	fileStoragePath, fileStoragePathExists := os.LookupEnv("FILE_STORAGE_PATH")
	if fileStoragePathExists {
		AppConfig.FileStoragePath = fileStoragePath
	} else {
		AppConfig.FileStoragePath = *flagFileStoragePath
	}
	databaseDSN, databaseDSNExists := os.LookupEnv("DATABASE_DSN")
	if databaseDSNExists {
		AppConfig.DatabaseDSN = databaseDSN
	} else {
		AppConfig.DatabaseDSN = *flagDatabaseDSN
	}
	secretKey, secretKeyExists := os.LookupEnv("SECRET_KEY")
	if secretKeyExists {
		AppConfig.SecretKey = secretKey
	} else if *flagSecretKey != "" {
		AppConfig.SecretKey = *flagSecretKey
	}
	auditFile, auditFileExists := os.LookupEnv("AUDIT_FILE")
	if auditFileExists {
		AppConfig.AuditFile = auditFile
	} else if *flagAuditFile != "" {
		AppConfig.AuditFile = *flagAuditFile
	}
	auditURL, auditURLExists := os.LookupEnv("AUDIT_URL")
	if auditURLExists {
		AppConfig.AuditURL = auditURL
	} else if *flagAuditURL != "" {
		AppConfig.AuditURL = *flagAuditURL
	}
}
