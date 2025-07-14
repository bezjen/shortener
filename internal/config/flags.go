package config

import "flag"

var FlagRunAddr string
var FlagShortURLAddr string

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "port to run server")
	flag.StringVar(&FlagShortURLAddr, "b", "http://localhost:8080", "address and port of tiny url")
	flag.Parse()
}
