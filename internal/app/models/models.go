package models

type RequestShotenerURL struct {
	URL string `json:"url"`
}

type ResponseShortURL struct {
	Result string `json:"result"`
}

func NewRequestShotenerURL(url string) *RequestShotenerURL {
	return &RequestShotenerURL{
		URL: url,
	}
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

type RequestShortenerURLArray []RequestShortenerURLBatch
type ResponseShortenerURLArray []ResponseShortenerURLBatch

func NewRequestShortenerURLBatch(id string, url string) *RequestShortenerURLBatch {
	return &RequestShortenerURLBatch{
		URL: url,
		ID:  id,
	}
}

func NewResponseShortenerURLBatch(id string, url string) *ResponseShortenerURLBatch {
	return &ResponseShortenerURLBatch{
		URL: url,
		ID:  id,
	}
}
