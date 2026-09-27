package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mark-chakravarthi/chirpy/internal/auth"
	"github.com/mark-chakravarthi/chirpy/internal/database"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type LoginResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type RefreshResponse struct {
	Token string `json:"token"`
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
		ID:          newUser.ID,
		Email:       newUser.Email,
		CreatedAt:   newUser.CreatedAt,
		UpdatedAt:   newUser.UpdatedAt,
		IsChirpyRed: newUser.IsChirpyRed,
	}

	respondWithJSON(w, 201, user)
}

func (a *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds *int   `json:"expires_in_seconds"`
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

	token, err := auth.MakeJWT(user.ID, a.secret, time.Hour)
	if err != nil {
		fmt.Printf("Error creating token: %s", err)
		respondWithError(w, 500, "Could not create token")
		return
	}

	refreshToken := auth.MakeRefreshToken()

	now := time.Now()
	_, err = a.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(time.Hour * 24 * 60),
		RevokedAt: sql.NullTime{},
	})
	if err != nil {
		respondWithError(w, 500, "Could not create refresh token")
		return
	}

	respondWithJSON(w, 200, LoginResponse{
		ID:           user.ID,
		Email:        user.Email,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (a *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	record, err := a.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if record.RevokedAt.Valid || time.Now().After(record.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "refresh token expired or revoked")
		return
	}

	token, err := auth.MakeJWT(record.UserID, a.secret, time.Hour)
	if err != nil {
		fmt.Printf("Error creating token: %s", err)
		respondWithError(w, 500, "Could not create token")
		return
	}

	respondWithJSON(w, 200, RefreshResponse{
		Token: token,
	})
}
func (a *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	_, err = a.db.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		Token:     refreshToken,
		RevokedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	w.WriteHeader(204)
}

func (a *apiConfig) handleUpdateUserCreds(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userId, err := auth.ValidateJWT(token, a.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	params := parameters{}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 400, "Invalid request body")
		return
	}

	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 400, "Error hashing password. User not created.")
		return
	}

	user, err := a.db.UpdateUserCreds(r.Context(), database.UpdateUserCredsParams{
		ID:             userId,
		Email:          params.Email,
		HashedPassword: hashed_password,
	})
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})

}
