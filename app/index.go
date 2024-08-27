package app

import "net/http"

// Index API
func Index(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{ "app": "tars-upload-gateway" }`))
}
