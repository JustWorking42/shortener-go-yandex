package configs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("SERVER_ADDRESS", "test_server_address")
	os.Setenv("RUN_ADDR", "test_run_address")
	os.Setenv("LOGGER_LEVEL", "test_logger_level")
	os.Setenv("FILE_STORAGE_PATH", "test_file_storage_path")
	os.Setenv("DATABASE_DSN", "test_database_dsn")
	os.Setenv("ENABLE_HTTPS", "true")
	os.Setenv("SSL_CERT_PATH", "test_ssl_cert_path")
	config, err := ParseFlags()

	assert.ErrorAs(t, err, &ErrParseConfigJson)
	assert.Equal(t, "test_server_address", config.ServerAdr)
	assert.Equal(t, "test_run_address", config.RedirectHost)
	assert.Equal(t, "test_logger_level", config.LogLevel)
	assert.Equal(t, "test_file_storage_path", config.FileStoragePath)
	assert.Equal(t, "test_database_dsn", config.DBAddress)
	assert.Equal(t, true, config.EnableHTTPS)
	assert.Equal(t, "test_ssl_cert_path", config.SSLCertPath)
}

func TestUpdateConfig(t *testing.T) {
	config1 := Config{
		SSLCertPath:     "",
		ServerAdr:       "",
		RedirectHost:    "",
		LogLevel:        "",
		FileStoragePath: "",
		DBAddress:       "addr1",
		EnableHTTPS:     false,
	}

	config2 := Config{
		SSLCertPath:     "path2",
		ServerAdr:       "addr2",
		RedirectHost:    "host2",
		LogLevel:        "level2",
		FileStoragePath: "path2",
		DBAddress:       "addr2",
		EnableHTTPS:     true,
	}

	config1.updateConfig(config2)

	assert.Equal(t, "path2", config1.SSLCertPath)
	assert.Equal(t, "addr2", config1.ServerAdr)
	assert.Equal(t, "host2", config1.RedirectHost)
	assert.Equal(t, "level2", config1.LogLevel)
	assert.Equal(t, "path2", config1.FileStoragePath)
	assert.Equal(t, "addr1", config1.DBAddress)
	assert.Equal(t, true, config1.EnableHTTPS)
}

func TestIsFileStorageEnabled(t *testing.T) {
	config := &Config{
		FileStoragePath: "test_file_storage_path",
	}

	assert.True(t, config.IsFileStorageEnabled())

	config.FileStoragePath = ""
	assert.False(t, config.IsFileStorageEnabled())
}

func TestCreateConfigFromFile(t *testing.T) {
	testConfig := Config{
		ServerAdr:       "test_server_address",
		RedirectHost:    "test_redirect_host",
		LogLevel:        "test_logger_level",
		FileStoragePath: "test_file_storage_path",
		DBAddress:       "test_database_dsn",
		EnableHTTPS:     true,
		SSLCertPath:     "test_ssl_cert_path",
		GRPCServerAdr:   "test_grpc_server_address",
		TrustedSubnet:   "test_trusted_subnet",
	}
	tempFile, err := os.CreateTemp("", "config.json")
	assert.NoError(t, err)
	_, err = tempFile.Write([]byte(testJson))
	assert.NoError(t, err)
	tempFile.Close()

	config, err := createConfigFromFile(tempFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, testConfig, config)
}

var testJson = `
{
	"server_address": "test_server_address",
	"redirect_host": "test_redirect_host",
	"log_level": "test_logger_level",
	"file_storage_path": "test_file_storage_path",
	"database_dsn": "test_database_dsn",
	"enable_https": true,
	"cert_path": "test_ssl_cert_path",
	"trusted_subnet": "test_trusted_subnet",
	"grpc_server_address": "test_grpc_server_address"
}
`
