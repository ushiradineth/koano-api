include .env
export $(shell sed 's/=.*//' .env)

DATABASE ?= "postgres://$(PG_USER):$(PG_PASSWORD)@$(PG_URL)/$(PG_DATABASE)?sslmode=$(PG_SSLMODE)"

db_up:
	migrate -path database/migration/ -database $(DATABASE) -verbose up

db_down:
	migrate -path database/migration/ -database $(DATABASE) -verbose down

db_fix:
	migrate -path database/migration/ -database $(DATABASE) force 1

gen_docs:
	swag fmt && swag init -g cmd/api/main.go

run: gen_docs
	go run cmd/api/main.go
