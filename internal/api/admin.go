package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func (cfg *Config) getFileServerHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)

	resBody := fmt.Sprintf(`<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>`, cfg.fileServerHits.Load())
	io.WriteString(w, resBody)
}

func (cfg *Config) resetMetrics(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("PLATFORM") != "dev" {
		respondWithError(w, 403, "Unauthorized to make this action")
		return
	}

	cfg.fileServerHits.Store(0)

	if err := cfg.db.DeleteUsers(r.Context()); err != nil {
		respondWithError(w, 500, "Error deleting users")
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, "OK")
}
