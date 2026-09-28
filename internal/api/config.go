package api

import (
	"net/http"
	"sync/atomic"

	"github.com/mark-chakravarthi/chirpy/internal/database"
)

type Config struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	secret         string
	polkaKey       string
}

func NewConfig(db *database.Queries, secret, polkaKey string) *Config {
	return &Config{
		db:       db,
		secret:   secret,
		polkaKey: polkaKey,
	}
}

func (cfg *Config) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/app/", cfg.middlewareMetricsInc(middlewareLog(http.StripPrefix("/app", http.FileServer(http.Dir("static"))))))

	mux.HandleFunc("GET /api/healthz", handleHealthz)

	mux.HandleFunc("GET /admin/metrics", cfg.getFileServerHits)
	mux.HandleFunc("POST /admin/reset", cfg.resetMetrics)

	mux.HandleFunc("POST /api/chirps", cfg.handleCreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handleGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handleGetChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handleDeleteChirp)

	mux.HandleFunc("POST /api/users", cfg.handleCreateUser)
	mux.HandleFunc("PUT /api/users", cfg.handleUpdateUserCreds)
	mux.HandleFunc("POST /api/login", cfg.handleLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handleRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handleRevoke)

	mux.HandleFunc("POST /api/polka/webhooks", cfg.handlePolkaWebhook)

	return mux
}
