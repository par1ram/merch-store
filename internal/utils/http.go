// internal/utils/http.go
package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// JSONResponse отправляет JSON-ответ клиенту с продвинутым логированием.
func JSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logrus.WithFields(logrus.Fields{
			"status":  status,
			"payload": payload,
		}).WithError(err).Error("failed to encode JSON response")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	logrus.WithFields(logrus.Fields{
		"status": status,
	}).Debug("JSON response sent successfully")
}

// ErrorResponse отправляет JSON-ответ с ошибкой и логирует отправку ошибки.
func JSONErrorResponse(w http.ResponseWriter, status int, message string) {
	logrus.WithFields(logrus.Fields{
		"status":  status,
		"message": message,
	}).Warn("sending error response")

	JSONResponse(w, status, map[string]string{"error": message})
}
