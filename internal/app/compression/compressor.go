// Package compression provides functionality for compressing HTTP requests.
package compression

import (
	"compress/gzip"
	"net/http"
)

// GzipRequestMiddleware is a middleware function that decompresses gzipped HTTP requests.
func GzipRequestMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			h(w, r)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "Unable to create gzip reader: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer gz.Close()

		r.Body = gz
		h(w, r)
	}
}
