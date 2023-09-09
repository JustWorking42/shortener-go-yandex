package handlers

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/urlgenerator"
)

const (
	incorectData   = "Incorrect Data"
	shortLinkRegex = "^[/][a-zA-Z]+$"
)

func Webhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetRequest(w, r)
	case http.MethodPost:
		handlePostRequest(w, r)
	default:
		sendError(w, incorectData, http.StatusBadRequest)
	}
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	match, err := regexp.MatchString(shortLinkRegex, r.URL.Path)

	if err != nil {
		sendError(w, incorectData, http.StatusBadRequest)
		return
	}

	if match {
		st := strings.Replace(r.URL.Path, "/", "", -1)
		link := (*storage.GetStorage())[st]

		if link == "" {
			sendError(w, incorectData, http.StatusBadRequest)
		}

		w.Header().Set("Location", link)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		sendError(w, incorectData, http.StatusBadRequest)
	}
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
