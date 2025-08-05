package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr      string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
}

var AppConfig Config

func ParseConfig() {
	flagServerAddr := flag.String("a", "localhost:8080", "port to run server")
	flagBaseURL := flag.String("b", "http://localhost:8080", "address and port of tiny url")
	flagLogLevel := flag.String("l", "info", "log level")
	flagFileStoragePath := flag.String("f", "./storage.json", "path to file with data")
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
}
