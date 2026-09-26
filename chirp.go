package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mark-chakravarthi/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 400, "Invalid request body")
		return
	}

	if params.Body == "" || params.UserID == uuid.Nil {
		respondWithError(w, 400, "Chirp/UserId cannot be empty")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	user, err := a.db.GetUser(r.Context(), params.UserID)
	if err != nil {
		respondWithError(w, 400, "User does not exist")
		return
	}

	chirpBody := removeProfanity(params.Body)

	now := time.Now()
	newChirp, err := a.db.CreateChirp(r.Context(), database.CreateChirpParams{
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

func (a *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	fetchedChirps, err := a.db.GetChirps(r.Context())
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

	respondWithJSON(w, 200, chirps)
}

func (a *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 400, "Invalid chirp ID")
		return
	}

	chirp, err := a.db.GetChirp(r.Context(), id)
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
