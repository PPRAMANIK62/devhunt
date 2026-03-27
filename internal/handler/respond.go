package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, map[string]any{"data": data})
}

func writeError(w http.ResponseWriter, err error) {
	status := apperr.HTTPStatus(err)

	msg := "an internal error occured"
	code := "INTERNAL_ERORR"

	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		code = string(appErr.Type)
		// For internal errors, don't expose details to client
		if status != http.StatusInternalServerError {
			msg = appErr.Message
		} else {
			// Log the real error internally
			slog.Error("internal error", "error", err)
		}
	}

	WriteJSON(w, status, map[string]string{"error": msg, "code": code})
}
