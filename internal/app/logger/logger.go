package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log zap.Logger = *zap.NewNop()

func Init(textLevel string) error {
	level, err := zap.ParseAtomicLevel(textLevel)
	if err != nil {
		return err
	}

	config := zap.NewDevelopmentConfig()
	config.Level = level
	logger, err := config.Build()
	if err != nil {
		return err
	}

	Log = *logger
	return nil
}

func RequestLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h(w, r)
		duration := time.Since(start)
		sugar := Log.Sugar()
		sugar.Infoln("uri", r.RequestURI, "method", r.Method, "duration", duration)
	}
}

func ResponseLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseData := &responseData{}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h(&lw, r)
		sugar := Log.Sugar()
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
