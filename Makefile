include .env
export $(shell sed 's/=.*//' .env)

DATABASE ?= "postgres://$(PG_USER):$(PG_PASSWORD)@$(PG_URL)/$(PG_DATABASE)?sslmode=$(PG_SSLMODE)"

db_up:
	migrate -path database/migration/ -database $(DATABASE) -verbose up

db_down:
	migrate -path database/migration/ -database $(DATABASE) -verbose down

db_fix:
	migrate -path database/migration/ -database $(DATABASE) force 1

db_seed:
	go run cmd/seeder/main.go

gen_docs:
	swag fmt && swag init -g cmd/api/main.go

run: gen_docs
	go run cmd/api/main.go

run_watch: gen_docs
	gow run cmd/api/main.go

build:
	go build -o cron-be cmd/api/main.go

build_image:
	docker build -t cron-be:dev -f deployments/Dockerfile .

compose_up:
	docker compose -f deployments/docker-compose.yml --env-file .env up -d

compose_down:
	docker compose -f deployments/docker-compose.yml --env-file .env down

test:
	go test -v -cover -failfast ./...

install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/mitranim/gow@latest
