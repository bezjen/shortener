package config

import "flag"

var FlagRunAddr string
var FlagShortURLArrd string

func parseFlags() {
	flag.StringVar(&FlagRunAddr, "a", "http://localhost:8000", "address and port to run server")
	flag.StringVar(&FlagShortURLArrd, "b", "http://localhost:8000", "address and port of tiny url")
	flag.Parse()
}
