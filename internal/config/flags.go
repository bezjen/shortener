package config

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
)

type Config struct {
	ServerAddr      string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
	DatabaseDSN     string
	SecretKey       []byte
}

var AppConfig Config

func ParseConfig() {
	flagServerAddr := flag.String("a", "localhost:8080", "port to run server")
	flagBaseURL := flag.String("b", "http://localhost:8080", "address and port of tiny url")
	flagLogLevel := flag.String("l", "info", "log level")
	flagFileStoragePath := flag.String("f", "", "path to file with data")
	flagDatabaseDSN := flag.String("d", "", "postgres data source name in format `postgres://username:password@host:port/database_name?sslmode=disable`")
	flagSecretKey := flag.String("s", "", "authorization secret key")
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
		AppConfig.SecretKey = []byte(secretKey)
	} else if *flagSecretKey != "" {
		AppConfig.SecretKey = []byte(*flagDatabaseDSN)
	} else {
		generatedKey, err := generateRandomKey(32)
		if err != nil {
			log.Fatal("Failed to generate secret key:", err)
		}
		AppConfig.SecretKey = generatedKey
	}
}

func generateRandomKey(length int) ([]byte, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	return key, nil
}
