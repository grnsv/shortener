package middleware

import (
	"net"
	"net/http"
)

// Internal returns a middleware that allows access only from the specified trusted subnet.
// Requests from outside the trusted subnet receive a 403 Forbidden response.
func Internal(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, subnet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			if !subnet.Contains(net.ParseIP(r.Header.Get("X-Real-IP"))) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
