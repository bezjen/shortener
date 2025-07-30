package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr      string `env:"SERVER_ADDRESS"`
	ShortURLAddr string `env:"BASE_URL"`
}

var FlagsConfig Config

func ParseFlags() {
	addr, addrExists := os.LookupEnv("SERVER_ADDRESS")
	if addrExists {
		FlagsConfig.RunAddr = addr
	} else {
		flag.StringVar(&FlagsConfig.RunAddr, "a", ":8080", "port to run server")
	}
	baseURL, baseURLExists := os.LookupEnv("BASE_URL")
	if baseURLExists {
		FlagsConfig.ShortURLAddr = baseURL
	} else {
		flag.StringVar(&FlagsConfig.ShortURLAddr, "b", "http://localhost:8080", "address and port of tiny url")
	}
	flag.Parse()
}
