package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/JustWorking42/shortener-go-yandex/internal/app"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/compression"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/logger"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/urlgenerator"

	"github.com/go-chi/chi/v5"
)

const (
	incorectData = "Incorrect Data"
)

func Webhook(app *app.App) *chi.Mux {

	router := chi.NewRouter()

	handleGetRequest := func(w http.ResponseWriter, r *http.Request) {
		handleGetRequest(app, w, r)
	}

	handlePostRequest := func(w http.ResponseWriter, r *http.Request) {
		handlePostRequest(app, w, r)
	}

	handleShortenPost := func(w http.ResponseWriter, r *http.Request) {
		handleShortenPost(app, w, r)
	}

	pingDB := func(w http.ResponseWriter, r *http.Request) {
		pingDB(app, w, r)
	}

	handleShortenPostArray := func(w http.ResponseWriter, r *http.Request) {
		handleShortenPostArray(app, w, r)
	}

	router.Get("/{id}", combinedMiddleware(handleGetRequest))

	router.Post("/", combinedMiddleware(handlePostRequest))

	router.Post("/api/shorten", combinedMiddleware(handleShortenPost))

	router.Get("/ping", logger.RequestLogging(logger.ResponseLogging(pingDB)))

	router.Post("/api/shorten/batch", combinedMiddleware(handleShortenPostArray))

	router.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		sendError(w, errors.New("MethodNotAllowed"), incorectData, http.StatusBadRequest)
	})
	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		sendError(w, errors.New("MethodNotFound"), incorectData, http.StatusBadRequest)
	})
	return router
}

func handleShortenPost(app *app.App, w http.ResponseWriter, r *http.Request) {

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

	err := app.Storage.Save(r.Context(), *savedURL)

	if err != nil {
		logger.Log.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := models.NewResponseShortURL(
		fmt.Sprintf("%s/%s", app.RedirectHost, shortID),
	)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}
}

func handleGetRequest(app *app.App, w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	savedURL, err := app.Storage.Get(r.Context(), id)

	if err != nil {
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", savedURL.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func handlePostRequest(app *app.App, w http.ResponseWriter, r *http.Request) {
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

	err = app.Storage.Save(r.Context(), *savedURL)

	if err != nil {
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", app.RedirectHost, shortID)))
}

func handleShortenPostArray(app *app.App, w http.ResponseWriter, r *http.Request) {

	var originalURLsArray models.RequestShortenerURLArray

	if err := json.NewDecoder(r.Body).Decode(&originalURLsArray); err != nil {
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	var shortURLsArray models.ResponseShortenerURLArray
	var shortURLsArraySave []storage.SavedURL

	for _, item := range originalURLsArray {
		if item.URL == "" {
			sendError(w, errors.New("IncorectBody"), incorectData, http.StatusBadRequest)
			return
		}
		shortURL := urlgenerator.CreateShortLink()

		shortURLsArray = append(
			shortURLsArray,
			*models.NewResponseShortenerURLBatch(item.ID, fmt.Sprintf("%s/%s", app.RedirectHost, shortURL)),
		)

		shortURLsArraySave = append(shortURLsArraySave, *storage.NewSavedURL(shortURL, item.URL))

	}

	err := app.Storage.SaveArray(r.Context(), shortURLsArraySave)

	if err != nil {
		logger.Log.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(shortURLsArray); err != nil {
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}
}

func pingDB(app *app.App, w http.ResponseWriter, r *http.Request) {
	err := app.Storage.Ping(r.Context())
	if err != nil {
		sendError(w, err, "", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func sendError(w http.ResponseWriter, err error, message string, statusCode int) {
	logger.Log.Sugar().Error(err)
	http.Error(w, message, statusCode)
}

func combinedMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return compression.GzipRequestMiddleware(
		logger.RequestLogging(
			logger.ResponseLogging(
				compression.GzipResponseMiddleware(h),
			),
		),
	)
}
