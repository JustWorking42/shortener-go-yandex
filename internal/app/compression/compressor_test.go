package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestGzipRequestMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(body)
	})

	gzipHandler := GzipRequestMiddleware(handler)

	server := httptest.NewServer(gzipHandler)
	defer server.Close()

	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	if _, err := gz.Write([]byte("test request body")); err != nil {
		t.Error(err)
	}
	if err := gz.Close(); err != nil {
		t.Error(err)
	}

	client := resty.New()
	resp, err := client.R().
		SetBody(buffer.Bytes()).
		SetHeader("Content-Encoding", "gzip").
		Post(server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test request body", resp.String())
}

func TestGzipRequestMiddleware_NoGzip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(body)
	})

	gzipHandler := GzipRequestMiddleware(handler)

	server := httptest.NewServer(gzipHandler)
	defer server.Close()

	client := resty.New()
	resp, err := client.R().
		SetBody([]byte("test request body")).
		Post(server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test request body", resp.String())
}

func BenchmarkGzipRequestMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
	})

	gzipHandler := GzipRequestMiddleware(handler)

	server := httptest.NewServer(gzipHandler)
	defer server.Close()

	client := resty.New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buffer bytes.Buffer
		gz := gzip.NewWriter(&buffer)
		_, _ = gz.Write([]byte("test request body"))
		_ = gz.Close()

		client.R().
			SetBody(buffer.Bytes()).
			SetHeader("Content-Encoding", "gzip").
			Post(server.URL)
	}
}
