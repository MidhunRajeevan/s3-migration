package util

import (
	"encoding/json"
	"net/http"

	"github.com/MidhunRajeevan/s3-migration/model"
)

// BadRequest response
func BadRequest(w *http.ResponseWriter, msg string) {
	error := model.Error{Code: 400, Message: msg}
	(*w).WriteHeader(http.StatusBadRequest)
	(*w).Header().Set("Content-Type", "application/json")
	json.NewEncoder(*w).Encode(error)
}

// NotFound response
func NotFound(w *http.ResponseWriter, msg string) {
	error := model.Error{Code: 404, Message: msg}
	(*w).WriteHeader(http.StatusNotFound)
	(*w).Header().Set("Content-Type", "application/json")
	json.NewEncoder(*w).Encode(error)
}

// MethodNotAllowed response
func MethodNotAllowed(w *http.ResponseWriter, msg string) {
	error := model.Error{Code: 405, Message: msg}
	(*w).WriteHeader(http.StatusMethodNotAllowed)
	(*w).Header().Set("Content-Type", "application/json")
	json.NewEncoder(*w).Encode(error)
}

// InternalServerError response
func InternalServerError(w *http.ResponseWriter, msg string) {
	error := model.Error{Code: 500, Message: msg}
	(*w).WriteHeader(http.StatusInternalServerError)
	(*w).Header().Set("Content-Type", "application/json")
	json.NewEncoder(*w).Encode(error)
}
