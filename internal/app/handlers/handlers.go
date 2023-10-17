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
	"go.uber.org/zap"

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

	router.Get("/{id}", combinedMiddleware(app.Logger, handleGetRequest))

	router.Post("/", combinedMiddleware(app.Logger, handlePostRequest))

	router.Post("/api/shorten", combinedMiddleware(app.Logger, handleShortenPost))

	router.Get("/ping", logger.RequestLogging(app.Logger, logger.ResponseLogging(app.Logger, pingDB)))

	router.Post("/api/shorten/batch", combinedMiddleware(app.Logger, handleShortenPostArray))

	router.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		sendError(w, errors.New("MethodNotAllowed"), incorectData, http.StatusBadRequest)
	})
	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		sendError(w, errors.New("MethodNotFound"), incorectData, http.StatusBadRequest)
	})
	return router
}

func handleShortenPost(app *app.App, w http.ResponseWriter, r *http.Request) {

	var originalURL models.RequestShotenerURL

	if err := json.NewDecoder(r.Body).Decode(&originalURL); err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if originalURL.URL == "" {
		app.Logger.Sugar().Error("Original URL is empty cheack request body")
		sendError(w, errors.New("IncorectBody"), incorectData, http.StatusBadRequest)
		return
	}

	link := originalURL.URL
	shortID := urlgenerator.CreateShortLink()

	savedURL := storage.NewSavedURL(shortID, link)

	conflictURL, err := app.Storage.Save(r.Context(), *savedURL)
	statusCode := http.StatusCreated

	if err != nil {
		app.Logger.Sugar().Error(err)
		if ok := errors.Is(err, storage.ErrURLConflict); ok {
			statusCode = http.StatusConflict
			shortID = conflictURL
		} else {
			app.Logger.Sugar().Error(err)
			sendError(w, err, incorectData, http.StatusBadRequest)
			return
		}
	}

	response := models.NewResponseShortURL(fmt.Sprintf("%s/%s", app.RedirectHost, shortID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}
}

func handleGetRequest(app *app.App, w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	savedURL, err := app.Storage.Get(r.Context(), id)

	if err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", savedURL.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func handlePostRequest(app *app.App, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		app.Logger.Sugar().Error("Original URL is empty cheack request body")
		sendError(w, errors.New("body lenth is 0"), incorectData, http.StatusBadRequest)
		return
	}

	link := string(body)
	shortID := urlgenerator.CreateShortLink()

	savedURL := storage.NewSavedURL(shortID, link)

	conflictURL, err := app.Storage.Save(r.Context(), *savedURL)
	statusCode := http.StatusCreated

	if err != nil {
		app.Logger.Sugar().Error(err)
		if ok := errors.Is(err, storage.ErrURLConflict); ok {
			statusCode = http.StatusConflict
			shortID = conflictURL
		} else {
			app.Logger.Sugar().Error(err)
			sendError(w, err, incorectData, http.StatusBadRequest)
			return
		}
	}

	response := fmt.Sprintf("%s/%s", app.RedirectHost, shortID)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(response))
}

func handleShortenPostArray(app *app.App, w http.ResponseWriter, r *http.Request) {

	var originalURLsSlice []models.RequestShortenerURLBatch

	if err := json.NewDecoder(r.Body).Decode(&originalURLsSlice); err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	var shortURLsSlice []models.ResponseShortenerURLBatch
	var shortURLsSliceSave []storage.SavedURL

	for _, item := range originalURLsSlice {
		if item.URL == "" {
			app.Logger.Sugar().Error("Original url is empty check request body")
			sendError(w, errors.New("IncorectBody"), incorectData, http.StatusBadRequest)
			return
		}
		shortURL := urlgenerator.CreateShortLink()

		shortURLsSlice = append(
			shortURLsSlice,
			*models.NewResponseShortenerURLBatch(item.ID, fmt.Sprintf("%s/%s", app.RedirectHost, shortURL)),
		)

		shortURLsSliceSave = append(shortURLsSliceSave, *storage.NewSavedURL(shortURL, item.URL))

	}

	err := app.Storage.SaveArray(r.Context(), shortURLsSliceSave)

	if err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(shortURLsSlice); err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}
}

func pingDB(app *app.App, w http.ResponseWriter, r *http.Request) {
	err := app.Storage.Ping(r.Context())
	if err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, "", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func sendError(w http.ResponseWriter, err error, message string, statusCode int) {
	http.Error(w, message, statusCode)
}

func combinedMiddleware(log *zap.Logger, h http.HandlerFunc) http.HandlerFunc {
	return compression.GzipRequestMiddleware(
		logger.RequestLogging(
			log,
			logger.ResponseLogging(
				log,
				compression.GzipResponseMiddleware(h),
			),
		),
	)
}
