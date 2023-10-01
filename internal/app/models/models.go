package models

type RequestShotenerURL struct {
	URL string `json:"url"`
}

type ResponseShortURL struct {
	Result string `json:"result"`
}
