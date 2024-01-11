// Package cookie provides functionality for handling cookies.
package cookie

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/JustWorking42/shortener-go-yandex/internal/app"
	"github.com/golang-jwt/jwt/v4"
)

var secretKey = []byte("secret-key")

// UserID is a type alias for string.
type UserID string

// handleError writes an error message to the HTTP response writer if there is an error.
func handleError(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}
}

// CookieCheckMiddleware is a middleware function that checks for a JWT token in the cookie and generate cookie.
func CookieCheckMiddleware(app *app.App, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwtToken")
		var userID string
		if errors.Is(err, http.ErrNoCookie) || err != nil {
			userID, err = app.UserManager.GenerateUserID(r.Context())
			handleError(w, err)
			cookie, err = createCookie(app, userID, r.Context())
			handleError(w, err)
		} else {
			jwtToken := cookie.Value
			userID, err = getUserID(jwtToken)
			if err != nil {
				userID, err = app.UserManager.GenerateUserID(r.Context())
				handleError(w, err)
				cookie, err = createCookie(app, userID, r.Context())
				handleError(w, err)
			}
		}
		http.SetCookie(w, cookie)
		ctx := context.WithValue(r.Context(), UserID("UserID"), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// OnlyAuthorizedMiddleware is a middleware function that checks if the user is authorized.
func OnlyAuthorizedMiddleware(app *app.App, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwtToken")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		jwtToken := cookie.Value
		_, err = getUserID(jwtToken)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// getUserID extracts the user ID from the JWT token.
func getUserID(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["userID"].(string)
		return userID, nil
	}

	return "", errors.New("Unauthorized")
}

// generateToken generates a JWT token for the given user ID.
func generateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// createCookie creates a new cookie with the JWT token.
func createCookie(app *app.App, userID string, ctx context.Context) (*http.Cookie, error) {
	jwtToken, err := generateToken(userID)
	if err != nil {
		return nil, errors.New("BadRequest")
	}

	cookie := &http.Cookie{
		Name:  "jwtToken",
		Value: jwtToken,
	}
	return cookie, nil
}
