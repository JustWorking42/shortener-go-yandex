package configs

import (
	"flag"
	"os"
)

type Config struct {
	ServerAdr       string
	RedirectHost    string
	LogLevel        string
	FileStoragePath string
}

func (c *Config) IsFileStorageEnabled() bool {
	return c.FileStoragePath != ""
}

var serverConfig Config

func GetServerConfig() *Config {
	return &serverConfig
}

func ParseFlags() error {
	os.Environ()
	flag.StringVar(&serverConfig.ServerAdr, "a", ":8080", "")
	flag.StringVar(&serverConfig.RedirectHost, "b", "http://localhost:8080", "")
	flag.StringVar(&serverConfig.LogLevel, "ll", "info", "")
	flag.StringVar(&serverConfig.FileStoragePath, "f", "/tmp/short-url-db.json", "")
	flag.Parse()

	if serverAdress := os.Getenv("SERVER_ADDRESS"); serverAdress != "" {
		serverConfig.ServerAdr = serverAdress
	}

	if redirectHost := os.Getenv("RUN_ADDR"); redirectHost != "" {
		serverConfig.RedirectHost = redirectHost
	}

	if logLevel := os.Getenv("LOGGER_LEVEL"); logLevel != "" {
		serverConfig.LogLevel = logLevel

	}

	if fileStoragePath, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		serverConfig.FileStoragePath = fileStoragePath

	}
	return nil
}
