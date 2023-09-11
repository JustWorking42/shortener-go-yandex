package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestWebhook(t *testing.T) {

	tests := []struct {
		name             string
		url              string
		preConfig        func()
		additionalAssert func(resp *resty.Response, expected string)
		expected         string
		statusCode       int
		methodType       string
		requestBody      []byte
	}{
		{
			name:       "GetFail",
			url:        "/fdfd",
			expected:   "Incorrect Data",
			statusCode: http.StatusBadRequest,
			additionalAssert: func(resp *resty.Response, expected string) {
				assert.Equal(t, expected, strings.TrimRight(string(resp.Body()), "\n"))
			},
			methodType: http.MethodGet,
		},
		{
			name:       "GetSuccess",
			url:        "/FHDds",
			expected:   "https://practicum.yandex.ru",
			statusCode: http.StatusTemporaryRedirect,
			additionalAssert: func(resp *resty.Response, expected string) {
				assert.Equal(t, expected, strings.TrimRight(resp.Header().Get("Location"), "\n"))
			},
			preConfig: func() {
				storage.Init()
				(*storage.GetStorage())["FHDds"] = "https://practicum.yandex.ru"
			},
			methodType: http.MethodGet,
		},
		{
			name:       "NotFound",
			url:        "/df/sa/fsdf/asd",
			expected:   "Incorrect Data",
			statusCode: http.StatusBadRequest,
			additionalAssert: func(resp *resty.Response, expected string) {
				assert.Equal(t, expected, strings.TrimRight(string(resp.Body()), "\n"))
			},
			methodType: http.MethodGet,
		},
		{
			name:       "MethodNotAllowed",
			url:        "/",
			expected:   "Incorrect Data",
			statusCode: http.StatusBadRequest,
			additionalAssert: func(resp *resty.Response, expected string) {
				assert.Equal(t, expected, strings.TrimRight(string(resp.Body()), "\n"))
			},
			methodType: http.MethodGet,
		},
		{
			name:       "PostSuccess",
			url:        "/",
			statusCode: http.StatusCreated,
			additionalAssert: func(resp *resty.Response, _ string) {
				assert.Equal(t, "text/plain", resp.Header().Get("Content-Type"))
				assert.Regexp(t, "^[/][a-zA-Z]+$", string(resp.Body()))
			},
			requestBody: []byte("https://practicum.yandex.ru"),
			methodType:  http.MethodPost,
			preConfig: func() {
				storage.Init()
			},
		},
		{
			name:       "PostFailEmptyBody",
			url:        "/",
			expected:   "Incorrect Data",
			statusCode: http.StatusBadRequest,
			additionalAssert: func(resp *resty.Response, expected string) {
				assert.Equal(t, expected, strings.TrimRight(string(resp.Body()), "\n"))
			},
			requestBody: []byte{},
			methodType:  http.MethodPost,
			preConfig: func() {
				storage.Init()
			},
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			if test.preConfig != nil {
				test.preConfig()
			}
			server := httptest.NewServer(Webhook())
			defer server.Close()

			client := resty.New()
			client.SetRedirectPolicy(resty.NoRedirectPolicy())

			var resp *resty.Response

			if test.methodType == http.MethodGet {
				resp, _ = client.R().Get(server.URL + test.url)
			} else if test.methodType == http.MethodPost {
				resp, _ = client.R().SetBody(test.requestBody).Post(server.URL + test.url)
			}

			assert.Equal(t, test.statusCode, resp.StatusCode())

			if test.additionalAssert != nil {
				test.additionalAssert(resp, test.expected)
			}
		})
	}
}
