package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestHandleGetRequestFaill(t *testing.T) {

	server := httptest.NewServer(Webhook())
	defer server.Close()

	client := resty.New()

	resp, _ := client.R().Get(server.URL + "/fdfd")

	expected := "Incorrect Data"

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
	assert.Equal(t, expected, strings.TrimRight(string(resp.Body()), "\n"))

}

func TestHandleGetRequestSuccess(t *testing.T) {
	expected := "https://practicum.yandex.ru"

	server := httptest.NewServer(Webhook())
	defer server.Close()

	storage.Init()
	(*storage.GetStorage())["FHDds"] = "https://practicum.yandex.ru"

	client := resty.New()

	client.SetRedirectPolicy(resty.NoRedirectPolicy())

	resp, _ := client.R().Get(server.URL + "/FHDds")

	fmt.Println(resp.StatusCode())

	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode())
	assert.Equal(t, expected, strings.TrimRight(resp.Header().Get("Location"), "\n"))
}
