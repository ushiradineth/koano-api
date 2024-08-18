# Go backend for Cron

- Check [main repo](https://github.com/ushiradineth/cron) for details about the project.

## Running the Project Locally

### Clone the repository

- `git clone https://github.com/ushiradineth/cron-be`

### Environment variables

- Check the `.env.example` file for the required environment variables.
- Use `cp .env.example .env` to create the `.env` file.

### Start the Postgres Database

- Run `docker compose -f deployments/docker-compose.yml up -d` to start the Postgres Database and Adminer.
- Wait for a moment for the database to initialize.

### Connect to the database

- You can use Adminer, a web-based administration tool included in the setup, to manage your database. Access Adminer at [localhost:9090](http://localhost:9090).
- In Adminer, use the following credentials:
  - Server: postgres:5432
  - Username: cron
  - Password: password
  - Database: cron

### Run Database Migrations

- Run `make db_up` to run the latest Database Migration.

### Install Go Watch

- `go install github.com/mitranim/gow@latest`

### Run the Go Server

- `go run cmd/api/main.go` or `gow run cmd/api/main.go`

### Run Tests

- `go test -v -cover ./...` or `gow test -v -cover ./...`

### Run the Seeder

- Note: Make sure the database is the development database
- `go run cmd/seeder/main.go`

## Build Docker Image

### Build the image

- Run `docker build -t cron-be:prod -f deployments/Dockerfile .` to build the image.

### Run the image using Docker Compose

- Uncomment the `cron-be` service in `docker-compose.yml`.
- Replace `PG_URL` in `.env` with `postgres:5432`
- Run `docker compose -f deployments/docker-compose.yml up -d` to start the Postgres Database, Adminer, and the Cron Go HTTP Server.
