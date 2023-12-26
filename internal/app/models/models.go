// Package models provides data structures and methods related to the application's data model.
package models

// RequestShotenerURL represents a request to shorten a URL.
type RequestShotenerURL struct {
	URL string `json:"url"`
}

// ResponseShortURL represents a shortened URL.
type ResponseShortURL struct {
	Result string `json:"result"`
}

// NewResponseShortURL creates a new ResponseShortURL instance.
func NewResponseShortURL(url string) *ResponseShortURL {
	return &ResponseShortURL{
		Result: url,
	}
}

// RequestShortenerURLBatch represents a batch request to shorten URLs.
type RequestShortenerURLBatch struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

// ResponseShortenerURLBatch represents a batch of shortened URLs.
type ResponseShortenerURLBatch struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}

// NewResponseShortenerURLBatch creates a new ResponseShortenerURLBatch instance.
func NewResponseShortenerURLBatch(id string, url string) *ResponseShortenerURLBatch {
	return &ResponseShortenerURLBatch{
		URL: url,
		ID:  id,
	}
}

// DeleteTask represents a task to delete a URL.
type DeleteTask struct {
	URL    string
	UserID string
}
