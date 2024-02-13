// Package handlers handles all incoming HTTP requests and routes them to the appropriate handler functions.
// It uses the chi router for routing and middleware for common tasks like logging and error handling.
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/JustWorking42/shortener-go-yandex/internal/app"
	accesscontrol "github.com/JustWorking42/shortener-go-yandex/internal/app/accessControl"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/compression"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/cookie"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/logger"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	incorectData = "Incorrect Data"
)

// Webhook sets up the HTTP server and defines the routes.
// It configures the router to use compression middleware for responses with specified content types.
func Webhook(app *app.App) *chi.Mux {

	router := chi.NewRouter()

	router.Use(middleware.Compress(5, "text/html", "text/plain", "application/json"))

	handleGetRequest := func(w http.ResponseWriter, r *http.Request) {
		HandleGetRequest(app, w, r)
	}

	handlePostRequest := func(w http.ResponseWriter, r *http.Request) {
		HandlePostRequest(app, w, r)
	}

	handleShortenPost := func(w http.ResponseWriter, r *http.Request) {
		HandleShortenPost(app, w, r)
	}

	pingDB := func(w http.ResponseWriter, r *http.Request) {
		PingDB(app, w, r)
	}

	handleShortenPostArray := func(w http.ResponseWriter, r *http.Request) {
		HandleShortenPostArray(app, w, r)
	}

	handleGetUserURLs := func(w http.ResponseWriter, r *http.Request) {
		HandleGetUserURLs(app, w, r)
	}

	handleDelete := func(w http.ResponseWriter, r *http.Request) {
		HandleDeleteURLs(app, w, r)
	}

	handleGetStats := func(w http.ResponseWriter, r *http.Request) {
		HandleGetStats(app, w, r)
	}

	router.Get("/{id}", combinedMiddleware(app, handleGetRequest))

	router.Post("/", combinedMiddleware(app, handlePostRequest))

	router.Post("/api/shorten", combinedMiddleware(app, handleShortenPost))

	router.Get("/ping", logger.RequestLogging(app.Logger, logger.ResponseLogging(app.Logger, pingDB)))

	router.Post("/api/shorten/batch", combinedMiddleware(app, handleShortenPostArray))

	router.Get("/api/user/urls", cookie.OnlyAuthorizedMiddleware(app, combinedMiddleware(app, handleGetUserURLs)))

	router.Delete("/api/user/urls", combinedMiddleware(app, handleDelete))

	router.Get("/api/internal/stats", accesscontrol.CidrAccessMiddleware(app, combinedMiddleware(app, handleGetStats)))

	router.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		sendError(w, errors.New("MethodNotAllowed"), incorectData, http.StatusBadRequest)
	})
	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		sendError(w, errors.New("MethodNotFound"), incorectData, http.StatusBadRequest)
	})
	return router
}

// HandleShortenPost handles POST requests to "/api/shorten".
func HandleShortenPost(app *app.App, w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(cookie.UserID("UserID")).(string)

	var originalURL models.RequestShotenerURL
	if err := json.NewDecoder(r.Body).Decode(&originalURL); err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if originalURL.URL == "" {
		app.Logger.Sugar().Error("Original URL is empty check request body")
		sendError(w, errors.New("IncorectBody"), incorectData, http.StatusBadRequest)
		return
	}

	statusCode := http.StatusCreated

	savedURL, err := app.Repository.SaveURL(r.Context(), originalURL.URL, userID)
	if err != nil {
		app.Logger.Sugar().Error(err)
		if savedURL != (storage.SavedURL{}) {
			statusCode = http.StatusConflict
		} else {
			sendError(w, err, incorectData, http.StatusBadRequest)
			return
		}
	}

	response := models.NewResponseShortURL(fmt.Sprintf("%s/%s", app.RedirectHost, savedURL.ShortURL))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}
}

func HandleGetRequest(app *app.App, w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	savedURL, err := app.Repository.GetURL(r.Context(), id)

	if err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	if savedURL.IsDeleted {
		w.WriteHeader(http.StatusGone)
		return
	}

	w.Header().Set("Location", savedURL.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// HandlePostRequest handles POST requests to "/".
func HandlePostRequest(app *app.App, w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(cookie.UserID("UserID")).(string)

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
		sendError(w, errors.New("body length is  0"), incorectData, http.StatusBadRequest)
		return
	}

	link := string(body)

	statusCode := http.StatusCreated

	savedURL, err := app.Repository.SaveURL(r.Context(), link, userID)
	if err != nil {
		app.Logger.Sugar().Error(err)
		if savedURL != (storage.SavedURL{}) {
			statusCode = http.StatusConflict
		} else {
			sendError(w, err, incorectData, http.StatusBadRequest)
			return
		}
	}

	response := fmt.Sprintf("%s/%s", app.RedirectHost, savedURL.ShortURL)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(response))
}

// HandleShortenPostArray handles POST requests to "/api/shorten/batch".
func HandleShortenPostArray(app *app.App, w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(cookie.UserID("UserID")).(string)
	var originalURLsSlice []models.RequestShortenerURLBatch

	if err := json.NewDecoder(r.Body).Decode(&originalURLsSlice); err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	savedURLsSlice, err := app.Repository.SaveURLArray(r.Context(), originalURLsSlice, userID)
	if err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(savedURLsSlice); err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, incorectData, http.StatusBadRequest)
	}
}

// PingDB checks the connection to the database.
func PingDB(app *app.App, w http.ResponseWriter, r *http.Request) {
	err := app.Repository.PingDB(r.Context())
	if err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, "", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// sendError sends an HTTP error response with the given status code and message.
func sendError(w http.ResponseWriter, err error, message string, statusCode int) {
	http.Error(w, message, statusCode)
}

// combinedMiddleware combines several middleware functions into one.
func combinedMiddleware(app *app.App, h http.HandlerFunc) http.HandlerFunc {
	return cookie.CookieCheckMiddleware(app, compression.GzipRequestMiddleware(
		logger.RequestLogging(
			app.Logger,
			logger.ResponseLogging(
				app.Logger,
				h,
			),
		),
	),
	)
}

// HandleGetUserURLs retrieves all URLs associated with a user.
func HandleGetUserURLs(app *app.App, w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(cookie.UserID("UserID")).(string)
	urls, err := app.Repository.GetUserURLs(r.Context(), userID)

	if err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, "Failed to get URLs", http.StatusBadRequest)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]map[string]string, len(urls))
	for i, url := range urls {
		response[i] = map[string]string{
			"short_url":    fmt.Sprintf("%s/%s", app.RedirectHost, url.ShortURL),
			"original_url": url.OriginalURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, "Failed to encode response", http.StatusBadRequest)
	}
}

// HandleDeleteURLs deletes specified URLs.
func HandleDeleteURLs(app *app.App, w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(cookie.UserID("UserID")).(string)
	var urls []string
	err := json.NewDecoder(r.Body).Decode(&urls)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = app.Repository.DeleteURLs(r.Context(), userID, urls)
	if err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, "Failed to delete URLs", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// HandleGetStats return statistics of urls and users.
func HandleGetStats(app *app.App, w http.ResponseWriter, r *http.Request) {
	stats, err := app.Repository.GetStats(r.Context())
	if err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		app.Logger.Sugar().Error(err)
		sendError(w, err, "Failed to encode response", http.StatusBadRequest)
	}
}
