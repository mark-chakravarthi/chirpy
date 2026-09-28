package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mark-chakravarthi/chirpy/internal/api"
	"github.com/mark-chakravarthi/chirpy/internal/database"
)

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("SECRET")
	polkaKey := os.Getenv("POLKA_KEY")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to DB")
	}

	cfg := api.NewConfig(database.New(db), secret, polkaKey)

	server := &http.Server{
		Addr:    ":8080",
		Handler: cfg.Routes(),
	}

	log.Fatal(server.ListenAndServe())
}
