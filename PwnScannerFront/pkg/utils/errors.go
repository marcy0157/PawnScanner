package utils

import (
	"encoding/json"
	"net/http"
)

// @Description ErrorResponse rappresenta la struttura di errore per le API.

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// WriteError scrive una risposta di errore standardizzata.
// Accetta il codice HTTP e un messaggio di errore, e restituisce una risposta JSON.
func WriteError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Message: message,
		Code:    code,
	})
}
