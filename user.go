package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mark-chakravarthi/chirpy/internal/auth"
	"github.com/mark-chakravarthi/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (a *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	param := parameters{}
	decoder := json.NewDecoder(r.Body)

	defer r.Body.Close()

	err := decoder.Decode(&param)
	if err != nil {
		fmt.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 400, "Invalid request body")
		return
	}

	if param.Email == "" {
		respondWithError(w, 400, "Email cannot be empty")
		return
	}

	hashedPassword, err := auth.HashPassword(param.Password)
	if err != nil {
		respondWithError(w, 400, "Error hashing password. User not created.")
		return
	}

	now := time.Now().UTC()
	newUser, err := a.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:             uuid.New(),
		Email:          param.Email,
		CreatedAt:      now,
		UpdatedAt:      now,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		fmt.Printf("Error creating user: %s", err)
		respondWithError(w, 500, "Could not create user")
		return
	}

	user := User{
		ID:        newUser.ID,
		Email:     newUser.Email,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}

	respondWithJSON(w, 201, user)
}

func (a *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}

	decoder := json.NewDecoder(r.Body)

	defer r.Body.Close()

	err := decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 400, "Invalid request body")
		return
	}

	user, err := a.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	isCorrectPassword, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !isCorrectPassword {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	respondWithJSON(w, 200, User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})

}
