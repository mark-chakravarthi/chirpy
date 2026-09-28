# Chirpy

A REST API for a Twitter-like service, built in Go with PostgreSQL. (boot.dev project)

## Requirements

- Go 1.27+
- PostgreSQL
- [goose](https://github.com/pressly/goose) (for migrations)
- [sqlc](https://sqlc.dev) (only needed if you change queries in `sql/queries`)

## Setup

1. Create a `.env` file in the project root:

   ```
   DB_URL=postgres://user:password@localhost:5432/chirpy?sslmode=disable
   SECRET=<random string for signing JWTs>
   POLKA_KEY=<api key for the Polka webhook>
   PLATFORM=dev
   ```

2. Run migrations:

   ```
   goose -dir sql/schema postgres "$DB_URL" up
   ```

3. Run the server (from the repo root):

   ```
   go run ./cmd/chirpy
   ```

   The server listens on `:8080`.

## Layout

- `cmd/chirpy` — entrypoint (env/DB setup, starts the HTTP server)
- `internal/api` — HTTP handlers, routing, and middleware
- `internal/auth` — password hashing, JWTs, refresh tokens
- `internal/database` — sqlc-generated database access (do not edit by hand)
- `sql/schema` — goose migrations
- `sql/queries` — SQL used to generate `internal/database`
- `static` — files served under `/app/`

## API

| Method | Path | Description |
|---|---|---|
| GET | `/api/healthz` | Health check |
| POST | `/api/users` | Create a user |
| PUT | `/api/users` | Update email/password (auth required) |
| POST | `/api/login` | Log in, returns access + refresh tokens |
| POST | `/api/refresh` | Exchange a refresh token for a new access token |
| POST | `/api/revoke` | Revoke a refresh token |
| POST | `/api/chirps` | Create a chirp (auth required) |
| GET | `/api/chirps` | List chirps (optional `?author_id=` filter, `?sort=asc\|desc`, default `asc`) |
| GET | `/api/chirps/{chirpID}` | Get a single chirp |
| DELETE | `/api/chirps/{chirpID}` | Delete a chirp (auth required, author only) |
| POST | `/api/polka/webhooks` | Polka webhook for Chirpy Red upgrades |
| GET | `/admin/metrics` | File server hit count |
| POST | `/admin/reset` | Reset metrics and delete all users (dev only, `PLATFORM=dev`) |

Auth-required endpoints expect `Authorization: Bearer <JWT>`.
