package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/logger"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/urlgenerator"
	"github.com/go-chi/chi/v5"
)

const (
	incorectData = "Incorrect Data"
)

func Webhook() *chi.Mux {

	router := chi.NewRouter()

	router.Get("/{id}", logger.RequestLogging(logger.ResponseLogging(handleGetRequest)))
	router.Post("/", logger.RequestLogging(logger.ResponseLogging(handlePostRequest)))
	router.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		sendError(w, incorectData, http.StatusBadRequest)
	})
	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		sendError(w, incorectData, http.StatusBadRequest)
	})
	return router
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	storageMap, err := storage.GetStorage()

	if err != nil {
		log.Print(err)
		sendError(w, incorectData, http.StatusBadRequest)
	}

	link, ok := (*storageMap)[id]

	if !ok {
		sendError(w, incorectData, http.StatusBadRequest)
	}

	w.Header().Set("Location", link)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sendError(w, incorectData, http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("Error closing response body")
		}
	}()

	if err != nil {
		sendError(w, incorectData, http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		sendError(w, incorectData, http.StatusBadRequest)
		return
	}

	link := string(body)
	shortID := urlgenerator.CreateShortLink()

	storageMap, err := storage.GetStorage()

	if err != nil {
		log.Print(err)
		sendError(w, incorectData, http.StatusBadRequest)
	}

	(*storageMap)[shortID] = link

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", configs.GetServerConfig().RedirectHost, shortID)))
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	http.Error(w, message, statusCode)
}
