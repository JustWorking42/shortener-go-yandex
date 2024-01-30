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

	config, err := ParseFlags()

	assert.Nil(t, err)
	assert.Equal(t, "test_server_address", config.ServerAdr)
	assert.Equal(t, "test_run_address", config.RedirectHost)
	assert.Equal(t, "test_logger_level", config.LogLevel)
	assert.Equal(t, "test_file_storage_path", config.FileStoragePath)
	assert.Equal(t, "test_database_dsn", config.DBAddress)
	assert.Equal(t, true, config.EnableHTTPS)
}

func TestIsFileStorageEnabled(t *testing.T) {
	config := &Config{
		FileStoragePath: "test_file_storage_path",
	}

	assert.True(t, config.IsFileStorageEnabled())

	config.FileStoragePath = ""
	assert.False(t, config.IsFileStorageEnabled())
}
