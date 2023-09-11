package configs

import (
	"flag"
)

type Config struct {
	ServerAdr    string
	RedirectHost string
}

var ServerConfig Config

func ParseFlags() {
	flag.StringVar(&ServerConfig.ServerAdr, "a", ":8080", "")
	flag.StringVar(&ServerConfig.RedirectHost, "b", "http://localhost:8080", "")
	flag.Parse()
}
