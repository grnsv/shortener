package middleware

import (
	"net/http"
)

func WithDefaults(h http.HandlerFunc) http.HandlerFunc {
	return WithLogging(WithCompressing(h))
}
