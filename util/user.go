package util

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/ushiradineth/cron-be/models"
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
			HTTPError(w, http.StatusBadRequest, "User not found", StatusFail)
			return nil

		}

		HTTPError(w, http.StatusInternalServerError, err.Error(), StatusError)
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
	accessToken, err := GetJWT(r)
	if err != nil {
		HTTPError(w, http.StatusBadRequest, err.Error(), StatusFail)
		return nil
	}

	JWT, err := ParseAccessToken(accessToken)
	if err != nil {
		HTTPError(w, http.StatusUnauthorized, "Access Token is invalid or expired", StatusFail)
		return nil
	}

	if JWT.StandardClaims.Valid() != nil {
		HTTPError(w, http.StatusUnauthorized, "Access Token is invalid or expired", StatusFail)
		return nil
	}

	user := GetUser(w, JWT.Id.String(), db)
	if err != nil {
		HTTPError(w, http.StatusBadRequest, err.Error(), StatusFail)
	}

	return user
}
