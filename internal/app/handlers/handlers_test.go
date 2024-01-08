package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JustWorking42/shortener-go-yandex/internal/app"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage/mocks"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func mockApp(t *testing.T, storage *mocks.MockStorage) *app.App {
	conf := configs.Config{
		ServerAdr:       "8080",
		RedirectHost:    "http://localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "/tmp/short-url-db.json",
		DBAddress:       "",
	}

	ctx := context.Background()

	app, err := app.CreateApp(ctx, conf)
	assert.NoError(t, err)

	app.Storage = storage

	return app
}

func TestGetFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Get(gomock.Any(), gomock.Any()).Return(storage.SavedURL{}, errors.New("error"))

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().Get(server.URL + "/nonexistent")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestGetSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockResponse := *storage.NewSavedURL("existent", "dsas", "asda")
	mockStorage.EXPECT().Get(gomock.Any(), "existent").Return(mockResponse, nil)

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	resp, _ := client.R().Get(server.URL + "/existent")

	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode())
	assert.Equal(t, "dsas", resp.Header().Get("Location"))
	assert.NotEqual(t, "gzip", resp.Header().Get("Content-Encoding"))
}

func TestGetSuccessGone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockResponse := *storage.NewSavedURL("existent", "dsas", "asda")
	mockResponse.IsDeleted = true
	mockStorage.EXPECT().Get(gomock.Any(), "existent").Return(mockResponse, nil)

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	resp, _ := client.R().Get(server.URL + "/existent")

	assert.Equal(t, http.StatusGone, resp.StatusCode())
	assert.NotEqual(t, "gzip", resp.Header().Get("Content-Encoding"))
}

func TestHandleShortenPostFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Save(gomock.Any(), gomock.Any()).Return("", errors.New("error"))

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().SetBody(`{"URL": "invalid"}`).Post(server.URL + "/api/shorten")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestHandleShortenPostSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Save(gomock.Any(), gomock.Any()).Return("shortID", nil)
	app := mockApp(t, mockStorage)
	server := httptest.NewServer(Webhook(app))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().SetBody(`{"URL": "https://valid.com"}`).Post(server.URL + "/api/shorten")

	assert.NoError(t, err)
	var response models.ResponseShortURL
	json.Unmarshal(resp.Body(), &response)
	assert.Regexp(t, "^"+app.RedirectHost+"/[a-zA-Z]+$", response.Result)
	assert.Equal(t, http.StatusCreated, resp.StatusCode())
}

func TestHandleShortenPostConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Save(gomock.Any(), gomock.Any()).Return("shortID", storage.ErrURLConflict)

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().SetBody(`{"URL": "https://valid.com"}`).Post(server.URL + "/api/shorten")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode())
}

func TestHandleShortenPostEmptyBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server := httptest.NewServer(Webhook(mockApp(t, mocks.NewMockStorage(ctrl))))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().Post(server.URL + "/api/shorten")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestHandlePostRequestFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Save(gomock.Any(), gomock.Any()).Return("", errors.New("error"))

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().SetBody(`invalid`).Post(server.URL + "/")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestHandlePostRequestSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Save(gomock.Any(), gomock.Any()).Return("shortID", nil)

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().SetBody("https://valid.com").Post(server.URL + "/")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode())
}

func TestHandlePostRequestConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Save(gomock.Any(), gomock.Any()).Return("shortID", storage.ErrURLConflict)

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().SetBody(`https://valid.com`).Post(server.URL + "/")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode())
}

func TestHandlePostRequestEmptyBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server := httptest.NewServer(Webhook(mockApp(t, mocks.NewMockStorage(ctrl))))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().Post(server.URL + "/")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestPingDBFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Ping(gomock.Any()).Return(errors.New("error"))

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().Get(server.URL + "/ping")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestPingDBSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Ping(gomock.Any()).Return(nil)

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().Get(server.URL + "/ping")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestHandleShortenPostArraySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().SaveArray(gomock.Any(), gomock.Any()).Return(nil)

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, _ := client.R().SetBody(`[{"correlation_id": "1", "original_url": "https://valid.com"}, {"correlation_id": "2", "original_url": "https://valid.com"}]`).Post(server.URL + "/api/shorten/batch")

	assert.Equal(t, http.StatusCreated, resp.StatusCode())

	var expectedMap, responseMap []map[string]interface{}
	expected := `[
        {
            "correlation_id":  "1",
            "short_url": "shortID"
        },
        {
            "correlation_id":  "2",
            "short_url": "shortID"
        }
    ]`
	err := json.Unmarshal([]byte(expected), &expectedMap)
	assert.NoError(t, err)
	err = json.Unmarshal(resp.Body(), &responseMap)
	assert.NoError(t, err)

	for i := range expectedMap {
		for k := range expectedMap[i] {
			_, ok := responseMap[i][k]
			assert.True(t, ok, "key %s is missing in %d", k, i)
		}
	}
}

func TestHandleShortenPostArrayFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().SaveArray(gomock.Any(), gomock.Any()).Return(errors.New("error"))

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().SetBody(`[{"correlation_id": "1", "original_url": "invalid"}, {"correlation_id": "2", "original_url": "invalid"}]`).Post(server.URL + "/api/shorten/batch")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestHandleShortenPostArrayConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().SaveArray(gomock.Any(), gomock.Any()).Return(storage.ErrURLConflict)

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, err := client.R().SetBody(`[{"correlation_id": "1", "original_url": "https://valid.com"}]`).Post(server.URL + "/api/shorten/batch")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestHandleGetUserURLsNoContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().GetByUser(gomock.Any(), gomock.Any()).Return([]storage.SavedURL{}, nil)

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, _ := client.R().
		SetCookie(&http.Cookie{
			Name:  "jwtToken",
			Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiJ1c2VyX2lkIn0.qgwhzie6gvs8BiiUfGSuODdJSr4cOmR7pggYrG3bT78",
		}).
		Get(server.URL + "/api/user/urls")
	assert.Equal(t, http.StatusNoContent, resp.StatusCode())
}

func TestHandleGetUserURLsHasContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().GetByUser(gomock.Any(), gomock.Any()).Return([]storage.SavedURL{
		{
			OriginalURL: "https://valid.com",
			ShortURL:    "shortURL",
			UserID:      "user_id",
		},
	}, nil)

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, _ := client.R().
		SetCookie(&http.Cookie{
			Name:  "jwtToken",
			Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiJ1c2VyX2lkIn0.qgwhzie6gvs8BiiUfGSuODdJSr4cOmR7pggYrG3bT78",
		}).
		Get(server.URL + "/api/user/urls")

	assert.Equal(t, http.StatusOK, resp.StatusCode())

	var responseMap []map[string]interface{}
	err := json.Unmarshal(resp.Body(), &responseMap)
	assert.NoError(t, err)

	expectedMap := []map[string]interface{}{
		{
			"original_url": "https://valid.com",
			"short_url":    "http://localhost:8080/shortURL",
		},
	}

	assert.Equal(t, expectedMap, responseMap)
}

func TestHandleDeleteURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	server := httptest.NewServer(Webhook(mockApp(t, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, _ := client.R().SetBody(`["url1", "url2"]`).Delete(server.URL + "/api/user/urls")

	assert.Equal(t, http.StatusAccepted, resp.StatusCode())
}
