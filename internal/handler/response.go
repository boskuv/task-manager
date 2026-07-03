package handler

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	writeJSON(w, status, payload)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
