package config

import "flag"

type Config struct {
	RunAddr      string
	ShortURLAddr string
}

var FlagsConfig Config

func ParseFlags() {
	flag.StringVar(&FlagsConfig.RunAddr, "a", ":8080", "port to run server")
	flag.StringVar(&FlagsConfig.ShortURLAddr, "b", "http://localhost:8080", "address and port of tiny url")
	flag.Parse()
}
