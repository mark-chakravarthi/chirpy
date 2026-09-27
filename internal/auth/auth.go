package auth

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"crypto/rand"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	matched, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return matched, err
	}

	return matched, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userID.String(),
	})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	userID, err := uuid.Parse(subject)
	if err != nil {
		return uuid.UUID{}, err
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authorizationHeader := headers.Get("Authorization")
	if authorizationHeader == "" {
		return "", errors.New("no authorization header included")
	}

	token, found := strings.CutPrefix(authorizationHeader, "Bearer ")
	if !found {
		return "", errors.New("malformed authorization header")
	}

	return strings.TrimSpace(token), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authorizationHeader := headers.Get("Authorization")
	if authorizationHeader == "" {
		return "", errors.New("no authorization header included")
	}

	key, found := strings.CutPrefix(authorizationHeader, "ApiKey ")
	if !found {
		return "", errors.New("malformed authorization header")
	}

	return strings.TrimSpace(key), nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	return hex.EncodeToString(key)
}
