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
