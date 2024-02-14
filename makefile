DATABASE ?= "mysql://cron:password@tcp(localhost:3306)/cron?multiStatements=true"

db_up: 
	migrate -path db/migration/ -database $(DATABASE) -verbose up

db_down: 
	migrate -path db/migration/ -database $(DATABASE) -verbose down

db_fix: 
	migrate -path db/migration/ -database $(DATABASE) force 1