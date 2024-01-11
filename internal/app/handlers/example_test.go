package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage/mocks"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
)

func ExampleHandleShortenPost() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Save(gomock.Any(), gomock.Any()).Return("", nil)
	server := httptest.NewServer(Webhook(mockApp(nil, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, _ := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"URL": "https://example.com"}`).
		Post(server.URL + "/api/shorten")

	fmt.Println(resp.StatusCode())

	// Output:
	// 201
}

func ExampleHandleGetRequest() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Get(gomock.Any(), "shortID").Return(storage.SavedURL{OriginalURL: "https://example.com"}, nil)
	server := httptest.NewServer(Webhook(mockApp(nil, mockStorage)))
	defer server.Close()

	client := resty.New().SetRedirectPolicy(resty.NoRedirectPolicy())

	resp, _ := client.R().Get(server.URL + "/shortID")

	fmt.Println(resp.StatusCode())
	fmt.Println(resp.Header().Get("Location"))

	// Output:
	// 307
	// https://example.com
}

func ExampleHandleShortenPostArray() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().SaveArray(gomock.Any(), gomock.Any()).Return(nil)
	server := httptest.NewServer(Webhook(mockApp(nil, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, _ := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`[{"correlation_id": "1", "original_url": "https://example1.com"}, {"correlation_id": "2", "original_url": "https://example2.com"}]`).
		Post(server.URL + "/api/shorten/batch")

	fmt.Println(resp.StatusCode())

	// Output:
	// 201
}

func ExampleHandleGetUserURLs() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().GetByUser(gomock.Any(), gomock.Any()).Return([]storage.SavedURL{
		{
			OriginalURL: "https://example.com",
			ShortURL:    "shortURL",
			UserID:      "user_id",
		},
	}, nil)
	server := httptest.NewServer(Webhook(mockApp(nil, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, _ := client.R().
		SetCookie(&http.Cookie{
			Name:  "jwtToken",
			Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiJ1c2VyX2lkIn0.qgwhzie6gvs8BiiUfGSuODdJSr4cOmR7pggYrG3bT78",
		}).
		Get(server.URL + "/api/user/urls")

	fmt.Println(resp.StatusCode())
	fmt.Println(resp.String())

	// Output:
	// 200
	// [{"original_url":"https://example.com","short_url":"http://localhost:8080/shortURL"}]
}

func ExampleHandleDeleteURLs() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockStorage(ctrl)
	server := httptest.NewServer(Webhook(mockApp(nil, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, _ := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`["shortURL1", "shortURL2"]`).
		Delete(server.URL + "/api/user/urls")

	fmt.Println(resp.StatusCode())

	// Output:
	// 202
}

func ExamplePingDB() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Ping(gomock.Any()).Return(nil)
	server := httptest.NewServer(Webhook(mockApp(nil, mockStorage)))
	defer server.Close()

	client := resty.New()
	resp, _ := client.R().Get(server.URL + "/ping")

	fmt.Println(resp.StatusCode())

	// Output:
	// 200
}

func ExampleHandlePostRequest() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().Save(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	app := mockApp(nil, mockStorage)
	server := httptest.NewServer(Webhook(app))
	defer server.Close()

	client := resty.New()
	resp, _ := client.R().
		SetHeader("Content-Type", "text/plain").
		SetBody("https://example.com").
		Post(server.URL + "/")

	fmt.Println(resp.StatusCode())

	// Output:
	// 201
}
