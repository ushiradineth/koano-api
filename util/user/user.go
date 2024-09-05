package user

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/ushiradineth/cron-be/models"
	"github.com/ushiradineth/cron-be/util/auth"
	"github.com/ushiradineth/cron-be/util/response"
)

func GetUser(w http.ResponseWriter, idOrEmail string, db *sqlx.DB) *models.User {
	user := models.User{}
	var query string
	var args []interface{}
	var email string

	id, err := uuid.Parse(idOrEmail)
	if err != nil {
		id = uuid.Nil
		email = idOrEmail
	}

	if id == uuid.Nil {
		query = "SELECT * FROM users WHERE email=$1"
		args = append(args, email)
	} else {
		query = "SELECT * FROM users WHERE id=$1"
		args = append(args, id)
	}

	err = db.Get(&user, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.GenericBadRequestError(w, fmt.Errorf("User not found"))
			return nil
		}

		response.GenericServerError(w, err)
		return nil
	}

	return &user
}

func DoesUserExist(idStr string, email string, db *sqlx.DB) (bool, int, error) {
	var id uuid.UUID

	if idStr != "" {
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return false, 0, err
		}

		id = parsedID
	}

	var userCount int
	var query string
	var args []interface{}

	if idStr != "" {
		query = "SELECT COUNT(*) FROM users WHERE id=$1 OR email=$2"
		args = append(args, id)
	} else {
		query = "SELECT COUNT(*) FROM users WHERE email=$1"
	}

	args = append(args, email)

	err := db.Get(&userCount, query, args...)
	if err != nil {
		return false, 0, err
	}

	return userCount != 0, userCount, nil
}

func GetUserFromJWT(r *http.Request, w http.ResponseWriter, db *sqlx.DB) *models.User {
	accessToken, err := auth.GetJWT(r)
	if err != nil {
		response.GenericBadRequestError(w, err)
		return nil
	}

	JWT := auth.ParseAccessToken(w, accessToken)
	if JWT == nil {
		return nil
	}

	user := GetUser(w, JWT.Id.String(), db)

	return user
}
