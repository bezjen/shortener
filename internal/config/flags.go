package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr string
	BaseURL    string
}

var AppConfig Config

func ParseConfig() {
	flagServerAddr := flag.String("a", "localhost:8080", "port to run server")
	flagBaseURL := flag.String("b", "http://localhost:8080", "address and port of tiny url")
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
}
