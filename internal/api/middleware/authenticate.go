package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/grnsv/shortener/internal/logger"
)

const cookieName = "token"

var signingMethod = jwt.SigningMethodHS256

type Claims struct {
	jwt.RegisteredClaims
}

type contextKey string

const UserIDContextKey contextKey = "userID"

func Authenticate(key string, logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string
			claims, err := getClaims(r, key)
			if err != nil {
				logger.Debug(err)
				userID, err = generateUserID()
				if err != nil {
					logger.Error(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				err = refreshCookie(w, key, userID)
				if err != nil {
					logger.Error(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				userID = claims.Subject
				if userID == "" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				err = refreshCookie(w, key, userID)
				if err != nil {
					logger.Error(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getClaims(r *http.Request, key string) (*Claims, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err
	}
	if err = cookie.Valid(); err != nil {
		return nil, err
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(cookie.Value, claims,
		func(t *jwt.Token) (interface{}, error) {
			if t.Method == nil || t.Method.Alg() != signingMethod.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(key), nil
		})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	return claims, nil
}

func refreshCookie(w http.ResponseWriter, key string, userID string) error {
	cookie, err := BuildAuthCookie(key, userID)
	if err != nil {
		return err
	}

	http.SetCookie(w, cookie)
	return nil
}

func BuildAuthCookie(key string, userID string) (*http.Cookie, error) {
	tokenString, err := buildJWTString(key, userID)
	if err != nil {
		return nil, err
	}
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(10 * 365 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	return cookie, nil
}

func buildJWTString(key string, userID string) (string, error) {
	token := jwt.NewWithClaims(signingMethod, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
	})

	return token.SignedString([]byte(key))
}

func generateUserID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
