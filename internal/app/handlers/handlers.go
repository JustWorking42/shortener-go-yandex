package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/compression"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/configs"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/logger"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/urlgenerator"

	"github.com/go-chi/chi/v5"
)

const (
	incorectData = "Incorrect Data"
)

func Webhook() *chi.Mux {

	router := chi.NewRouter()
	router.Get(
		"/{id}",
		compression.GzipRequestMiddleware(
			logger.RequestLogging(
				logger.ResponseLogging(
					compression.GzipResponseMiddleware(
						handleGetRequest,
					),
				),
			),
		),
	)
	router.Post(
		"/",
		compression.GzipRequestMiddleware(
			logger.RequestLogging(
				logger.ResponseLogging(
					compression.GzipResponseMiddleware(
						handlePostRequest,
					),
				),
			),
		),
	)
	router.Post(
		"/api/shorten",
		compression.GzipRequestMiddleware(
			logger.RequestLogging(
				logger.ResponseLogging(
					compression.GzipResponseMiddleware(
						handleShortenPost,
					),
				),
			),
		),
	)
	router.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		sendError(w, errors.New("MethodNotAllowed"), incorectData, http.StatusBadRequest)
	})
	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		sendError(w, errors.New("MethodNotFound"), incorectData, http.StatusBadRequest)
	})
	return router
}

func handleShortenPost(w http.ResponseWriter, r *http.Request) {

	var shortURL models.RequestShotenerURL

	if err := json.NewDecoder(r.Body).Decode(&shortURL); err != nil {
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if shortURL.URL == "" {
		sendError(w, errors.New("IncorectBody"), incorectData, http.StatusBadRequest)
		return
	}

	link := shortURL.URL
	shortID := urlgenerator.CreateShortLink()

	savedURL := storage.NewSavedURL(shortID, link)

	err := storage.Save(*savedURL)

	if err != nil {
		logger.Log.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := models.NewResponseShortURL(
		fmt.Sprintf("%s/%s", configs.GetServerConfig().RedirectHost, shortID),
	)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	savedURL, err := storage.Get(id)

	if err != nil {
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", savedURL.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		sendError(w, errors.New("body lenth is 0"), incorectData, http.StatusBadRequest)
		return
	}

	link := string(body)
	shortID := urlgenerator.CreateShortLink()

	savedURL := storage.NewSavedURL(shortID, link)

	err = storage.Save(*savedURL)

	if err != nil {
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", configs.GetServerConfig().RedirectHost, shortID)))
}

func sendError(w http.ResponseWriter, err error, message string, statusCode int) {
	logger.Log.Sugar().Error(err)
	http.Error(w, message, statusCode)
}
