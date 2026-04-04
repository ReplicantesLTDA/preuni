package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const headerRequestID = "X-Request-ID"

// RequestID injects a unique request ID into the response header.
// If the incoming request already carries X-Request-ID it is preserved.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(headerRequestID)
		if id == "" {
			id = newRequestID()
			r.Header.Set(headerRequestID, id)
		}
		w.Header().Set(headerRequestID, id)
		next.ServeHTTP(w, r)
	})
}

func newRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
