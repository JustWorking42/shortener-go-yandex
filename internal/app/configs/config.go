// Package configs provides functionality for parsing and managing application configurations.
package configs

import (
	"flag"
	"os"
	"strconv"
)

// Config holds the application's configuration.
type Config struct {
	ServerAdr       string
	RedirectHost    string
	LogLevel        string
	FileStoragePath string
	DBAddress       string
	EnableHTTPS     bool
	CertPath        string
}

// IsFileStorageEnabled checks if file storage is enabled.
func (c *Config) IsFileStorageEnabled() bool {
	return c.FileStoragePath != ""
}

// ParseFlags parses command-line flags and environment variables to populate the Config structure.
func ParseFlags() (*Config, error) {
	serverConfig := Config{}
	os.Environ()
	flag.StringVar(&serverConfig.ServerAdr, "a", ":8080", "Serer address")
	flag.StringVar(&serverConfig.RedirectHost, "b", "http://localhost:8080", "Redirection host")
	flag.StringVar(&serverConfig.LogLevel, "ll", "info", "Loglevel")
	flag.StringVar(&serverConfig.FileStoragePath, "f", "/tmp/short-url-db.json", "File storage path")
	flag.StringVar(&serverConfig.DBAddress, "d", "", "DB address")
	flag.BoolVar(&serverConfig.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&serverConfig.CertPath, "c", "", "Cert path")

	flag.Parse()

	if certPath := os.Getenv("SSL_CERT_PATH"); certPath != "" {
		serverConfig.CertPath = certPath
	}

	if enableHTTPS := os.Getenv("ENABLE_HTTPS"); enableHTTPS != "" {
		serverConfig.EnableHTTPS, _ = strconv.ParseBool(enableHTTPS)
	}

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

	if dbStoragePath, exist := os.LookupEnv("DATABASE_DSN"); exist {
		serverConfig.DBAddress = dbStoragePath
	}
	return &serverConfig, nil
}
