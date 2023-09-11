package configs

import (
	"flag"
	"os"
)

type Config struct {
	ServerAdr    string
	RedirectHost string
}

var serverConfig Config

func GetServerConfig() *Config {
	return &serverConfig
}

func ParseFlags() error {
	os.Environ()
	flag.StringVar(&serverConfig.ServerAdr, "a", ":8080", "")
	flag.StringVar(&serverConfig.RedirectHost, "b", "http://localhost:8080", "")
	flag.Parse()

	if serverAdress := os.Getenv("SERVER_ADDRESS"); serverAdress != "" {
		serverConfig.ServerAdr = serverAdress
	}

	if redirectHost := os.Getenv("RUN_ADDR"); redirectHost != "" {
		serverConfig.RedirectHost = redirectHost
	}
	return nil
}
