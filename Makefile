DATABASE ?= "postgres://cron:password@localhost:5432/cron?sslmode=disable"

db_up:
	migrate -path database/migration/ -database $(DATABASE) -verbose up

db_down:
	migrate -path database/migration/ -database $(DATABASE) -verbose down

db_fix:
	migrate -path database/migration/ -database $(DATABASE) force 1

gen_docs:
	swag init -g cmd/api/main.go
