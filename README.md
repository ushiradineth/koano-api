# Go backend for Cron

- Check [main repo](https://github.com/ushiradineth/cron) for details about the project.

## Running the Project Locally

### Clone the repository

- `git clone https://github.com/ushiradineth/cron-be`

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
  - Username: cron
  - Password: password
  - Database: cron

### Install Go Migrate

- `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

### Run Database Migrations

- Run `make db_up` to run the latest Database Migration.

### Run the Seeder

- Note: Make sure the database is the development database
- `go run cmd/seeder/main.go`

### Install Go Watch (optional)

- `go install github.com/mitranim/gow@latest`

### Run the Go Server

- `go run cmd/api/main.go` or `gow run cmd/api/main.go`

## Build the Cron API

### Build the image

- Run `docker build -t cron-be:go -f deployments/Dockerfile .` to build the image.

### Run the image using Docker Compose

- Uncomment the `cron-be` service in `docker-compose.yml`.
- Run `docker compose -f deployments/docker-compose.yml --env-file .env up -d` to start the Postgres Database, Adminer, and the Cron Go HTTP Server.

## Testing the Project

### Run the tests

- `go test -v -cover -failfast test ./...`
