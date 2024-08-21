package util

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/ushiradineth/cron-be/models"
)

func GetUser(idOrEmail string, db *sqlx.DB) (*models.User, error) {
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
			return nil, errors.New("User not found")
		}
		return nil, err
	}

	return &user, nil
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

func GetUserFromJWT(r *http.Request, db *sqlx.DB) (*models.User, int, error) {
	accessToken, err := GetJWT(r)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	JWT, err := ParseAccessToken(accessToken)
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("Access Token is invalid or expired")
	}

	if JWT.StandardClaims.Valid() != nil {
		return nil, http.StatusUnauthorized, errors.New("Access Token is invalid or expired")
	}

	user, err := GetUser(JWT.Id.String(), db)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New(fmt.Sprint(err))
	}

	return user, http.StatusOK, nil
}
