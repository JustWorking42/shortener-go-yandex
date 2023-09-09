package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func TestHandleGetRequestFaill(t *testing.T) {
	expected := "Incorrect Data"
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(handleGetRequest)

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, recorder.Code, http.StatusBadRequest)
	assert.Equal(t, strings.TrimRight(recorder.Body.String(), "\n"), expected)

}

func TestHandleGetRequestSuccess(t *testing.T) {
	expected := "localHost:8080"
	storage.Init()
	(*storage.GetStorage())["FHDds"] = "localHost:8080"

	req, err := http.NewRequest("GET", "/FHDds", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(handleGetRequest)

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
	assert.Equal(t, expected, strings.TrimRight(recorder.Header()["Location"][0], "\n"))
}
