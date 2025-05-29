# Koano API (Go Backend)

[![License: BSL](https://img.shields.io/badge/license-BSL--1.1-blue.svg)](LICENSE)

This is the Go backend powering [Koano](https://koano.app) — a modern, Zen-inspired calendar and scheduling engine.

> For the frontend/client, see the [koano frontend repository](https://github.com/ushiradineth/koano).

---

## Running the Project Locally

### Clone the repository

- `git clone https://github.com/ushiradineth/koano-api`

### Install binaries

#### Required

- `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

#### Optional

- `go install github.com/swaggo/swag/cmd/swag@latest`
- `go install github.com/mitranim/gow@latest`

### Environment variables

- Check the `.env.example` file for the required environment variables.
- Use `cp .env.example .env` to create the `.env` file.

### Start the Postgres Database

- Run `docker compose -f deployments/docker-compose.yml --env-file .env up -d` to start the Postgres Database and Adminer.
- Wait for a moment for the database to initialize.

### Connect to the database

- You can use Adminer, a web-based administration tool included in the setup, to manage your database. Access Adminer at [localhost:9090](http://localhost:9090).
- In Adminer, use the following credentials:
  - Server: postgres:5432
  - Username: koano
  - Password: password
  - Database: koano

### Run Database Migrations

- Run `make db_up` to run the latest Database Migration.

### Run the Seeder

- Note: Make sure the database is the development database
- `go run cmd/seeder/main.go` or `make db_seed`

### Run the Go Server

- `go run cmd/api/main.go` or `make run`
- `gow run cmd/api/main.go` or `make run_watch`

## Build the Koano API

### Build the image

- Run `docker build -t koano-api:go -f deployments/Dockerfile .` or `make build_image` to build the image.

### Run the image using Docker Compose

- Uncomment the `koano-api` service in `docker-compose.yml`.
- Run `docker compose -f deployments/docker-compose.yml --env-file .env up -d` or `make compose_up` to start the Postgres Database, Adminer, and the Koano Go HTTP Server.

## Testing the Project

### Run the tests

- `go test -v -cover -failfast test ./...` or `make test`
