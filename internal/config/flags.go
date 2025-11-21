// Package config provides configuration management for the URL shortening service.
// It handles command-line flags, environment variables, and configuration defaults.
package config

import (
	"encoding/json"
	"flag"
	"os"
)

// Config holds all application configuration settings.
// Settings can be provided via command-line flags, environment variables, or a configuration file.
type Config struct {
	ServerAddr      string `json:"server_address"`    // Server address in format "host:port"
	BaseURL         string `json:"base_url"`          // Base URL for shortened links
	LogLevel        string `json:"log_level"`         // Logging level (debug, info, warn, error)
	FileStoragePath string `json:"file_storage_path"` // Path to file storage (if using file backend)
	DatabaseDSN     string `json:"database_dsn"`      // PostgreSQL connection string
	SecretKey       string `json:"secret_key"`        // JWT secret key for authentication
	AuditFile       string `json:"audit_file"`        // Path to audit log file
	AuditURL        string `json:"audit_url"`         // URL for remote audit logging
	EnableHTTPS     bool   `json:"enable_https"`      // Enable HTTPS flag
}

// AppConfig is the global application configuration instance.
var AppConfig Config

// ParseConfig parses command-line flags and environment variables to populate AppConfig.
// Priority (highest to lowest):
// 1. Environment variables
// 2. Command-line flags
// 3. Configuration file (JSON)
// 4. Default values
func ParseConfig() {
	// Define flags with empty defaults to distinguish between "not set" and "set to default"
	flagServerAddr := flag.String("a", "", "port to run server")
	flagBaseURL := flag.String("b", "", "address and port of tiny url")
	flagLogLevel := flag.String("l", "", "log level")
	flagFileStoragePath := flag.String("f", "", "path to file with data")
	flagDatabaseDSN := flag.String("d", "", "postgres data source name")
	flagSecretKey := flag.String("k", "", "authorization secret key")
	flagAuditFile := flag.String("audit-file", "", "path to audit file")
	flagAuditURL := flag.String("audit-url", "", "audit url")
	flagEnableHTTPS := flag.String("s", "", "enable https")

	// New flags for configuration file
	flagConfig := flag.String("config", "", "path to config file")
	flagC := flag.String("c", "", "path to config file (shorthand)")

	flag.Parse()

	// Initialize with hardcoded defaults
	AppConfig = Config{
		ServerAddr:  "localhost:8080",
		BaseURL:     "http://localhost:8080",
		LogLevel:    "info",
		EnableHTTPS: false,
	}

	// Determine config file path: Env > Flag -config > Flag -c
	cfgPath := ""
	if envCfg, ok := os.LookupEnv("CONFIG"); ok {
		cfgPath = envCfg
	} else if *flagConfig != "" {
		cfgPath = *flagConfig
	} else if *flagC != "" {
		cfgPath = *flagC
	}

	// Load JSON config if path is provided
	if cfgPath != "" {
		fileData, err := os.ReadFile(cfgPath)
		if err == nil {
			// Use a struct with pointers to distinguish between missing fields and zero values
			type fileConfig struct {
				ServerAddr      *string `json:"server_address"`
				BaseURL         *string `json:"base_url"`
				LogLevel        *string `json:"log_level"`
				FileStoragePath *string `json:"file_storage_path"`
				DatabaseDSN     *string `json:"database_dsn"`
				SecretKey       *string `json:"secret_key"`
				AuditFile       *string `json:"audit_file"`
				AuditURL        *string `json:"audit_url"`
				EnableHTTPS     *bool   `json:"enable_https"`
			}
			var fc fileConfig
			if err := json.Unmarshal(fileData, &fc); err == nil {
				if fc.ServerAddr != nil {
					AppConfig.ServerAddr = *fc.ServerAddr
				}
				if fc.BaseURL != nil {
					AppConfig.BaseURL = *fc.BaseURL
				}
				if fc.LogLevel != nil {
					AppConfig.LogLevel = *fc.LogLevel
				}
				if fc.FileStoragePath != nil {
					AppConfig.FileStoragePath = *fc.FileStoragePath
				}
				if fc.DatabaseDSN != nil {
					AppConfig.DatabaseDSN = *fc.DatabaseDSN
				}
				if fc.SecretKey != nil {
					AppConfig.SecretKey = *fc.SecretKey
				}
				if fc.AuditFile != nil {
					AppConfig.AuditFile = *fc.AuditFile
				}
				if fc.AuditURL != nil {
					AppConfig.AuditURL = *fc.AuditURL
				}
				if fc.EnableHTTPS != nil {
					AppConfig.EnableHTTPS = *fc.EnableHTTPS
				}
			}
		}
	}

	// Apply Flags (override JSON/Default if set)
	if *flagServerAddr != "" {
		AppConfig.ServerAddr = *flagServerAddr
	}
	if *flagBaseURL != "" {
		AppConfig.BaseURL = *flagBaseURL
	}
	if *flagLogLevel != "" {
		AppConfig.LogLevel = *flagLogLevel
	}
	if *flagFileStoragePath != "" {
		AppConfig.FileStoragePath = *flagFileStoragePath
	}
	if *flagDatabaseDSN != "" {
		AppConfig.DatabaseDSN = *flagDatabaseDSN
	}
	if *flagSecretKey != "" {
		AppConfig.SecretKey = *flagSecretKey
	}
	if *flagAuditFile != "" {
		AppConfig.AuditFile = *flagAuditFile
	}
	if *flagAuditURL != "" {
		AppConfig.AuditURL = *flagAuditURL
	}
	if *flagEnableHTTPS != "" {
		AppConfig.EnableHTTPS = *flagEnableHTTPS == "true"
	}

	// Apply Environment Variables (override everything)
	if addr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		AppConfig.ServerAddr = addr
	}
	if baseURL, ok := os.LookupEnv("BASE_URL"); ok {
		AppConfig.BaseURL = baseURL
	}
	if logLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		AppConfig.LogLevel = logLevel
	}
	if fileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		AppConfig.FileStoragePath = fileStoragePath
	}
	if databaseDSN, ok := os.LookupEnv("DATABASE_DSN"); ok {
		AppConfig.DatabaseDSN = databaseDSN
	}
	if secretKey, ok := os.LookupEnv("SECRET_KEY"); ok {
		AppConfig.SecretKey = secretKey
	}
	if auditFile, ok := os.LookupEnv("AUDIT_FILE"); ok {
		AppConfig.AuditFile = auditFile
	}
	if auditURL, ok := os.LookupEnv("AUDIT_URL"); ok {
		AppConfig.AuditURL = auditURL
	}
	if enableHTTPS, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		AppConfig.EnableHTTPS = enableHTTPS == "true"
	}
}
