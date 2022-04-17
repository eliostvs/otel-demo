package web

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	resp, err := json.Marshal(data)
	if err != nil {
		ServerErrorResponse(w, err)
		return
	}

	// Append a newline to the JSON. This is just a small nicety to make it easier to view in terminal applications.
	resp = append(resp, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(resp)
}

func ErrorResponse(w http.ResponseWriter, status int, message interface{}, err error) {
	if WriteJSON(w, status, Envelope{"error": message, "cause": err.Error()}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func ServerErrorResponse(w http.ResponseWriter, err error) {
	message := "the server encountered a problem and could not process your request"
	ErrorResponse(w, http.StatusInternalServerError, message, err)
}
