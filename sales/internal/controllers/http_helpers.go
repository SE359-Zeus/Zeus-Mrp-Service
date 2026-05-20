package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func readJSON(r *http.Request, target any) error {
	return json.NewDecoder(r.Body).Decode(target)
}

func parseID(path string, prefix string) (uuid.UUID, bool) {
	idPart := strings.TrimPrefix(path, prefix)
	idPart = strings.Trim(idPart, "/")
	if idPart == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(idPart)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
