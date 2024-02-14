module cron

go 1.22.0

require (
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/joho/godotenv v1.5.1
)

require event v1.0.0

replace event => ./event

require db v1.0.0

replace db => ./db
