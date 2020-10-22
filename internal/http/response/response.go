package response

import (
	"encoding/json"
	"net/http"

	"n.eko.moe/neko/internal/http/endpoint"
)

// JSON encodes data to rw in JSON format. Returns a pointer to a
// HandlerError if encoding fails.
func JSON(w http.ResponseWriter, data interface{}, status int) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return &endpoint.HandlerError{
			Status:  http.StatusInternalServerError,
			Message: "unable to write JSON response",
			Err:     err,
		}
	}

	return nil
}

// Empty merely sets the response code to NoContent (204).
func Empty(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
