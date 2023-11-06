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

type UserID string

func CookieCheckMiddleware(app *app.App, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwtToken")

		var userID string
		if errors.Is(err, http.ErrNoCookie) {
			userID, err = app.UserManager.GenerateUserID(r.Context())
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			jwtToken, err := generateToken(userID)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			cookie := &http.Cookie{
				Name:  "jwtToken",
				Value: jwtToken,
			}
			http.SetCookie(w, cookie)
		} else {
			jwtToken := cookie.Value
			userID, err = getUserID(r.Context(), jwtToken)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		ctx := context.WithValue(r.Context(), UserID("UserID"), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func getUserID(ctx context.Context, tokenString string) (string, error) {
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
