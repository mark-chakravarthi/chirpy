package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mark-chakravarthi/chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	secret         string
	polkaKey       string
}

func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (a *apiConfig) getFileServerHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	w.WriteHeader(200)

	resBody := fmt.Sprintf(`<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>`, a.fileServerHits.Load())

	io.WriteString(w, resBody)
}

func (a *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	platform := os.Getenv("PLATFORM")

	if platform != "dev" {
		respondWithError(w, 403, "Unauthorized to make this action")
		return
	}

	a.fileServerHits.Store(0)

	err := a.db.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, "Error deleting users")
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, "OK")
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("SECRET")
	polkaKey := os.Getenv("POLKA_KEY")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to DB")
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()

	apiCfg := &apiConfig{
		fileServerHits: atomic.Int32{},
		db:             dbQueries,
		secret:         secret,
		polkaKey:       polkaKey,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(middlewareLog(http.StripPrefix("/app", http.FileServer(http.Dir(""))))))

	mux.HandleFunc("GET /api/healthz", handleHealthz)

	mux.HandleFunc("GET /admin/metrics", apiCfg.getFileServerHits)

	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)

	mux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)

	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)

	mux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirps)

	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handleGetChirp)

	mux.HandleFunc("POST /api/login", apiCfg.handleLogin)

	mux.HandleFunc("POST /api/refresh", apiCfg.handleRefresh)

	mux.HandleFunc("POST /api/revoke", apiCfg.handleRevoke)

	mux.HandleFunc("PUT /api/users", apiCfg.handleUpdateUserCreds)

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handleDeleteChirp)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlePolkaWebhook)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
