package seeder

import (
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func CreateUser(db *sqlx.DB) uuid.UUID {
	userId := uuid.New()

	_, err := db.Exec("INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4)", userId, faker.Name(), faker.Email(), faker.Password())
	if err != nil {
		panic(err)
	}

	return userId
}
