package middleware

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/preuni/pkg/errors"
)

// JSON writes v as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a standard error envelope response.
func writeError(w http.ResponseWriter, ae *apperrors.AppError) {
	type envelope struct {
		Error *apperrors.AppError `json:"error"`
	}
	JSON(w, ae.HTTPStatus(), envelope{Error: ae})
}

// ErrorResponse converts any error to a JSON error response.
// If err is not an *AppError, it is treated as an internal error.
func ErrorResponse(w http.ResponseWriter, err error) {
	ae := apperrors.As(err)
	if ae == nil {
		ae = apperrors.Internal(err)
	}
	writeError(w, ae)
}
