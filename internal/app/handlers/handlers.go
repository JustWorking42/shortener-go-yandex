package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/urlgenerator"
	"github.com/go-chi/chi/v5"
)

const (
	incorectData   = "Incorrect Data"
	shortLinkRegex = "^[/][a-zA-Z]+$"
)

func Webhook() *chi.Mux {

	router := chi.NewRouter()

	router.Get("/{id}", handleGetRequest)
	router.Post("/", handlePostRequest)
	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		sendError(w, incorectData, http.StatusBadRequest)
	})
	return router
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	link := (*storage.GetStorage())[id]

	if link == "" {
		sendError(w, incorectData, http.StatusBadRequest)
	}

	w.Header().Set("Location", link)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(link))
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sendError(w, incorectData, http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)

	if err != nil {
		sendError(w, incorectData, http.StatusBadRequest)
		return
	}
	link := string(body)
	shortID := urlgenerator.CreateShortLink()
	(*storage.GetStorage())[shortID] = link
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("http://localhost:8080/%s", shortID)))
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	http.Error(w, message, statusCode)
}
