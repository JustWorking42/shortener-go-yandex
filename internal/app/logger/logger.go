// Package logger provides functionality for logging HTTP requests and responses.
package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// CreateLogger creates a new zap logger with the specified log level.
func CreateLogger(textLevel string) (*zap.Logger, error) {
	level, err := zap.ParseAtomicLevel(textLevel)
	if err != nil {
		return nil, err
	}

	config := zap.NewDevelopmentConfig()
	config.Level = level
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// RequestLogging wraps an http handler function with logging functionality.
// Logs the request URI, method, and duration.
func RequestLogging(logger *zap.Logger, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h(w, r)
		duration := time.Since(start)
		logger.Info("request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
		)
	}
}

// ResponseLogging wraps an http handler function with logging functionality.
// Logs the response status and size.
func ResponseLogging(logger *zap.Logger, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseData := &responseData{}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h(&lw, r)
		logger.Info("response",
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
		)
	}
}

// responseData holds the status and size of an HTTP response.
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter is a custom response writer that logs the response status and size.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write writes the data to the underlying response writer and updates the response size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader writes the header to the underlying response writer and updates the response status.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
