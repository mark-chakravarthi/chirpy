package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mark-chakravarthi/chirpy/internal/auth"
	"github.com/mark-chakravarthi/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cfg *Config) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	user, err := cfg.db.GetUser(r.Context(), userId)
	if err != nil {
		respondWithError(w, 400, "User does not exist")
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	params := parameters{}

	err = decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 400, "Invalid request body")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	chirpBody := removeProfanity(params.Body)

	now := time.Now()
	newChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		Body:      chirpBody,
		UserID:    user.ID,
		CreatedAt: now,
		UpdatedAt: now,
	})

	chirp := Chirp{
		ID:        newChirp.ID,
		Body:      newChirp.Body,
		UserID:    newChirp.UserID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
	}

	respondWithJSON(w, 201, chirp)
}

var profaneWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func removeProfanity(s string) string {
	words := strings.Fields(s)

	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := profaneWords[loweredWord]; ok {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

func (cfg *Config) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	authorIDParam := r.URL.Query().Get("author_id")

	var authorID uuid.NullUUID
	if authorIDParam != "" {
		parsed, parseErr := uuid.Parse(authorIDParam)
		if parseErr != nil {
			respondWithError(w, 400, "Invalid author ID")
			return
		}
		authorID = uuid.NullUUID{UUID: parsed, Valid: true}
	}

	fetchedChirps, err := cfg.db.GetChirps(r.Context(), authorID)
	if err != nil {
		respondWithError(w, 400, "Error fetching chirps")
		return
	}

	chirps := []Chirp{}

	for _, chirp := range fetchedChirps {
		chirps = append(chirps, Chirp{
			ID:        chirp.ID,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		})
	}

	if r.URL.Query().Get("sort") == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	}

	respondWithJSON(w, 200, chirps)
}

func (cfg *Config) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 400, "Invalid chirp ID")
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		respondWithError(w, 404, "Error fetching chirp")
		return
	}

	c := Chirp{
		ID:        chirp.ID,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
	}

	respondWithJSON(w, 200, c)
}

func (cfg *Config) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 400, "Invalid chirp ID")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 404, "Chirp not found")
		return
	}

	if chirp.UserID != userId {
		respondWithError(w, 403, "You are not the author of this chirp")
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 500, "Error deleting chirp")
		return
	}

	w.WriteHeader(204)
}
