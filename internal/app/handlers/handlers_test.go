package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
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
		headers          map[string]string
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
			preConfig: func() {
				err := storage.Init()
				assert.NoError(t, err)
			},
		},
		{
			name:       "GetSuccess",
			url:        "/FHDds",
			expected:   "https://practicum.yandex.ru",
			statusCode: http.StatusTemporaryRedirect,
			additionalAssert: func(resp *resty.Response, expected string) {
				assert.Equal(t, expected, strings.TrimRight(resp.Header().Get("Location"), "\n"))
				assert.Equal(t, "", resp.Header().Get("Content-Encoding"))
			},
			preConfig: func() {
				err := storage.Init()
				assert.NoError(t, err)

				storageMap, err := storage.GetStorage()
				assert.NoError(t, err)

				(*storageMap)["FHDds"] = "https://practicum.yandex.ru"
			},
			methodType: http.MethodGet,
		},
		{
			name:       "GetSuccessGzip",
			url:        "/FHDds",
			expected:   "https://practicum.yandex.ru",
			statusCode: http.StatusTemporaryRedirect,
			additionalAssert: func(resp *resty.Response, expected string) {
				assert.Equal(t, expected, strings.TrimRight(resp.Header().Get("Location"), "\n"))
				assert.Equal(t, "gzip", resp.Header().Get("Content-Encoding"))
			},
			headers: map[string]string{
				"Accept-Encoding": "gzip",
			},
			preConfig: func() {
				err := storage.Init()
				assert.NoError(t, err)

				storageMap, err := storage.GetStorage()
				assert.NoError(t, err)

				(*storageMap)["FHDds"] = "https://practicum.yandex.ru"
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
				assert.Equal(t, "", resp.Header().Get("Content-Encoding"))
				assert.Regexp(t, "^[/][a-zA-Z]+$", string(resp.Body()))
			},
			requestBody: []byte("https://practicum.yandex.ru"),
			methodType:  http.MethodPost,
			preConfig: func() {
				err := storage.Init()
				assert.NoError(t, err)
			},
		},
		{
			name:       "PostSuccessGzip",
			url:        "/",
			statusCode: http.StatusCreated,
			additionalAssert: func(resp *resty.Response, _ string) {
				assert.Equal(t, "text/plain", resp.Header().Get("Content-Type"))
				assert.Equal(t, "gzip", resp.Header().Get("Content-Encoding"))
				assert.Regexp(t, "^[/][a-zA-Z]+$", string(resp.Body()))
			},
			requestBody: []byte("https://practicum.yandex.ru"),
			headers: map[string]string{
				"Accept-Encoding": "gzip",
			},
			methodType: http.MethodPost,
			preConfig: func() {
				err := storage.Init()
				assert.NoError(t, err)
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
				err := storage.Init()
				assert.NoError(t, err)
			},
		},
		{
			name:       "ShortenPostSuccessGzip",
			url:        "/api/shorten",
			statusCode: http.StatusCreated,
			additionalAssert: func(resp *resty.Response, _ string) {
				assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
				assert.Equal(t, "gzip", resp.Header().Get("Content-Encoding"))
				var response models.ResponseShortURL
				json.Unmarshal(resp.Body(), &response)
				assert.Regexp(t, "^"+configs.GetServerConfig().RedirectHost+"/[a-zA-Z]+$", response.Result)
			},
			requestBody: []byte(`{"URL": "https://practicum.yandex.ru"}`),
			methodType:  http.MethodPost,
			headers: map[string]string{
				"Accept-Encoding": "gzip",
			},
			preConfig: func() {
				err := configs.ParseFlags()
				assert.NoError(t, err)
				err = storage.Init()
				assert.NoError(t, err)
			},
		},
		{
			name:       "ShortenPostSucces",
			url:        "/api/shorten",
			statusCode: http.StatusCreated,
			additionalAssert: func(resp *resty.Response, _ string) {
				assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
				assert.Equal(t, "", resp.Header().Get("Content-Encoding"))
				var response models.ResponseShortURL
				json.Unmarshal(resp.Body(), &response)
				assert.Regexp(t, "^"+configs.GetServerConfig().RedirectHost+"/[a-zA-Z]+$", response.Result)
			},
			requestBody: []byte(`{"URL": "https://practicum.yandex.ru"}`),
			methodType:  http.MethodPost,

			preConfig: func() {
				err := storage.Init()
				assert.NoError(t, err)
			},
		},

		{
			name:       "ShortenPostFailInvalidBody",
			url:        "/api/shorten",
			expected:   "Incorrect Data",
			statusCode: http.StatusBadRequest,
			additionalAssert: func(resp *resty.Response, expected string) {
				assert.Equal(t, expected, strings.TrimRight(string(resp.Body()), "\n"))
			},
			requestBody: []byte(`{"invalid": "invalid"}`),
			methodType:  http.MethodPost,
			preConfig: func() {
				err := storage.Init()
				assert.NoError(t, err)
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
				resp, _ = client.R().SetHeaders(test.headers).Get(server.URL + test.url)
			} else if test.methodType == http.MethodPost {
				resp, _ = client.R().SetBody(test.requestBody).SetHeaders(test.headers).Post(server.URL + test.url)
			}

			assert.Equal(t, test.statusCode, resp.StatusCode())

			if test.additionalAssert != nil {
				test.additionalAssert(resp, test.expected)
			}
		})
	}
}
