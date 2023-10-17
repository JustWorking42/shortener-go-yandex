package models

type RequestShotenerURL struct {
	URL string `json:"url"`
}

type ResponseShortURL struct {
	Result string `json:"result"`
}

func NewResponseShortURL(url string) *ResponseShortURL {
	return &ResponseShortURL{
		Result: url,
	}
}

type RequestShortenerURLBatch struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type ResponseShortenerURLBatch struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}

func NewResponseShortenerURLBatch(id string, url string) *ResponseShortenerURLBatch {
	return &ResponseShortenerURLBatch{
		URL: url,
		ID:  id,
	}
}
