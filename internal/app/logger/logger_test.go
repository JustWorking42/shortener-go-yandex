package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestCreateLogger(t *testing.T) {
	logger, err := CreateLogger("debug")
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	assert.True(t, logger.Core().Enabled(zap.DebugLevel))
}

func TestRequestLogging(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	loggingHandler := RequestLogging(logger, handler)
	loggingHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Check if log was written
	assert.Equal(t, 1, recorded.Len())
	entry := recorded.All()[0]

	assert.Contains(t, entry.ContextMap(), "uri")
	assert.Contains(t, entry.ContextMap(), "method")
	assert.Contains(t, entry.ContextMap(), "duration")
}

func TestResponseLogging(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	loggingHandler := ResponseLogging(logger, handler)
	loggingHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Check if log was written
	assert.Equal(t, 1, recorded.Len())
	entry := recorded.All()[0]

	assert.Contains(t, entry.ContextMap(), "status")
	assert.Contains(t, entry.ContextMap(), "size")
}

func TestLoggingResponseWriter_Write(t *testing.T) {
	rr := httptest.NewRecorder()
	lrw := &loggingResponseWriter{
		ResponseWriter: rr,
		responseData:   &responseData{},
	}

	testSize, err := lrw.Write([]byte("OK"))
	assert.NoError(t, err)
	assert.Equal(t, testSize, lrw.responseData.size)
}

func TestLoggingResponseWriter_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	lrw := &loggingResponseWriter{
		ResponseWriter: rr,
		responseData:   &responseData{},
	}

	lrw.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusOK, lrw.responseData.status)
}
