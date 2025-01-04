package user

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/ushiradineth/cron-be/models"
	"github.com/ushiradineth/cron-be/util/auth"
	"github.com/ushiradineth/cron-be/util/response"
)

func GetUserById(id string, db *sqlx.DB) (*models.User, error) {
	user := models.User{}

	err := db.Get(&user, "SELECT * FROM users WHERE id=$1 AND active=true", id)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByEmail(email string, db *sqlx.DB) (*models.User, error) {
	user := models.User{}

	err := db.Get(&user, "SELECT * FROM users WHERE email=$1 AND active=true", email)
	if err != nil {
		return nil, err
	}

	return &user, nil
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

	user, err := GetUserById(JWT.Id.String(), db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.GenericBadRequestError(w, fmt.Errorf("User by id %s not found", JWT.Id.String()))
			return nil
		}

		response.GenericBadRequestError(w, err)
		return nil
	}

	return user
}

func IsEmailInUse(email string, id string, db *sqlx.DB) (bool, error) {
	var count int

	// Not checking active=true since deleted users can be restored
	err := db.Get(&count, "SELECT COUNT(*) FROM users WHERE email=$1 AND id!=$2", email, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}
