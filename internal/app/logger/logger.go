package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

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

func RequestLogging(logger *zap.Logger, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h(w, r)
		duration := time.Since(start)
		sugar := logger.Sugar()
		sugar.Infoln("uri", r.RequestURI, "method", r.Method, "duration", duration)
	}
}

func ResponseLogging(logger *zap.Logger, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseData := &responseData{}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h(&lw, r)
		sugar := logger.Sugar()
		sugar.Infoln("status", responseData.status, "size", responseData.size)
	}
}

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
