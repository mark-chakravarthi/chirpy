package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mark-chakravarthi/chirpy/internal/auth"
)

func (cfg *Config) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.polkaKey {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err = decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 400, "Invalid request body")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	_, err = cfg.db.UpgradeUserToChirpyRed(r.Context(), params.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, 404, "User not found")
			return
		}
		respondWithError(w, 500, "Error upgrading user")
		return
	}

	w.WriteHeader(204)
}
