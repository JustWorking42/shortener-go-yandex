// Package configs provides functionality for parsing and managing application configurations.
package configs

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"strconv"
)

// Config holds the application's configuration.
type Config struct {
	ServerAdr       string `json:"server_address"`
	RedirectHost    string `json:"redirect_host"`
	LogLevel        string `json:"log_level"`
	FileStoragePath string `json:"file_storage_path"`
	DBAddress       string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	SSLCertPath     string `json:"cert_path"`
}

// ErrParseConfigJson is returned when the config file cant be parsed.
var ErrParseConfigJson = errors.New("ErrParseConfigJson")

// IsFileStorageEnabled checks if file storage is enabled.
func (c *Config) IsFileStorageEnabled() bool {
	return c.FileStoragePath != ""
}

// ParseFlags parses command-line flags and environment variables to populate the Config structure.
func ParseFlags() (*Config, error) {
	serverConfig := Config{}
	var jsonConfigPath string
	os.Environ()
	flag.StringVar(&serverConfig.ServerAdr, "a", ":8080", "Serer address")
	flag.StringVar(&serverConfig.RedirectHost, "b", "http://localhost:8080", "Redirection host")
	flag.StringVar(&serverConfig.LogLevel, "ll", "info", "Loglevel")
	flag.StringVar(&serverConfig.FileStoragePath, "f", "/tmp/short-url-db.json", "File storage path")
	flag.StringVar(&serverConfig.DBAddress, "d", "", "DB address")
	flag.BoolVar(&serverConfig.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&serverConfig.SSLCertPath, "cr", "", "Cert path")
	flag.StringVar(&jsonConfigPath, "c", "", "JSON config")
	flag.Parse()

	if jsonPath := os.Getenv("CONFIG"); jsonPath != "" {
		jsonConfigPath = jsonPath
	}

	if certPath := os.Getenv("SSL_CERT_PATH"); certPath != "" {
		serverConfig.SSLCertPath = certPath
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

	jsonConfig, err := createConfigFromFile(jsonConfigPath)

	if err != nil {
		return &serverConfig, err
	}

	serverConfig.updateConfig(jsonConfig)
	return &serverConfig, nil
}

// createConfigFromFile reads a JSON config file and returns a Config structure.
func createConfigFromFile(configPath string) (Config, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, ErrParseConfigJson
	}
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return Config{}, ErrParseConfigJson
	}

	return config, nil
}

// updateConfig updates a Config structure by filling in any empty fields using another Config structure.
func (c *Config) updateConfig(config Config) {
	if c.SSLCertPath == "" && config.SSLCertPath != "" {
		c.SSLCertPath = config.SSLCertPath
	}
	if c.ServerAdr == "" && config.ServerAdr != "" {
		c.ServerAdr = config.ServerAdr
	}
	if c.RedirectHost == "" && config.RedirectHost != "" {
		c.RedirectHost = config.RedirectHost
	}
	if c.LogLevel == "" && config.LogLevel != "" {
		c.LogLevel = config.LogLevel
	}
	if c.FileStoragePath == "" && config.FileStoragePath != "" {
		c.FileStoragePath = config.FileStoragePath
	}
	if c.DBAddress == "" && config.DBAddress != "" {
		c.DBAddress = config.DBAddress
	}
	if !c.EnableHTTPS && config.EnableHTTPS {
		c.EnableHTTPS = config.EnableHTTPS
	}
}
