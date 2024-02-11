// Package configs provides functionality for parsing and managing application configurations.
package configs

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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
	TrustedSubnet   string `json:"trusted_subnet"`
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
	flag.StringVar(&serverConfig.TrustedSubnet, "t", "", "Trusted subnet")
	flag.Parse()

	if jsonPath, exist := os.LookupEnv("CONFIG"); exist {
		jsonConfigPath = jsonPath
	}

	if certPath, exist := os.LookupEnv("SSL_CERT_PATH"); exist {
		serverConfig.SSLCertPath = certPath
	}

	if enableHTTPS, exist := os.LookupEnv("ENABLE_HTTPS"); exist {
		if value, err := strconv.ParseBool(enableHTTPS); err == nil {
			serverConfig.EnableHTTPS = value
		}
	}

	if serverAdress, exist := os.LookupEnv("SERVER_ADDRESS"); exist {
		serverConfig.ServerAdr = serverAdress
	}

	if redirectHost, exist := os.LookupEnv("RUN_ADDR"); exist {
		serverConfig.RedirectHost = redirectHost
	}

	if logLevel, exist := os.LookupEnv("LOGGER_LEVEL"); exist {
		serverConfig.LogLevel = logLevel
	}

	if fileStoragePath, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		serverConfig.FileStoragePath = fileStoragePath
	}

	if dbStoragePath, exist := os.LookupEnv("DATABASE_DSN"); exist {
		serverConfig.DBAddress = dbStoragePath
	}

	if trustedSubnet, exit := os.LookupEnv("TRUSTED_SUBNET"); exit {
		serverConfig.TrustedSubnet = trustedSubnet
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
		return Config{}, fmt.Errorf("%w: unable to read config file: %v", ErrParseConfigJson, err)
	}
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return Config{}, fmt.Errorf("%w: unable to unmarshal config file: %v", ErrParseConfigJson, err)
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
