package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// RequestID ensures each request has a stable correlation ID.
// If the request already has X-Request-ID, it is preserved.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			var b [16]byte
			if _, err := rand.Read(b[:]); err != nil {
				// Fallback to a basic ID if crypto/rand fails (extremely rare)
				id = "fallback-id"
			} else {
				id = hex.EncodeToString(b[:])
			}
			r.Header.Set("X-Request-ID", id)
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}
