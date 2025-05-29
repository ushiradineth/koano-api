package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ushiradineth/koano-api/models"
	"github.com/ushiradineth/koano-api/util/auth"
	logger "github.com/ushiradineth/koano-api/util/log"
	"github.com/ushiradineth/koano-api/util/response"
	"github.com/ushiradineth/koano-api/util/user"
)

type API struct {
	db        *sqlx.DB
	validator *validator.Validate
	log       *logger.Logger
}

func New(db *sqlx.DB, validator *validator.Validate, log *logger.Logger) *API {
	return &API{
		db:        db,
		validator: validator,
		log:       log,
	}
}

// @Summary		Get User
// @Description	Get authenticated user based on the JWT sent with the request
// @Tags			User
// @Produce		json
// @Param			Path	path		UserPathParams	true	"UserPathParams"
// @Success		200		{object}	response.Response{data=models.User}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Security		BearerAuth
// @Router			/users/{user_id} [get]
func (api *API) Get(w http.ResponseWriter, r *http.Request) {
	path := UserPathParams{
		UserID: r.PathValue("user_id"),
	}

	if err := api.validator.Struct(path); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	user := user.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	if user.ID.String() != path.UserID {
		response.GenericUnauthenticatedError(w)
		return
	}

	user.Password = "redacted"

	api.log.Info.Printf("User %s has been retrieved", user.ID)

	response.HTTPResponse(w, user)
}

// @Summary		Create User
// @Description	Create User with the parameters sent with the request
// @Tags			User
// @Accept	  json
// @Produce		json
// @Param			Body	body		PostBodyParams	true	"PostBodyParams"
// @Success		200		{object}	response.Response{data=models.User}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Router			/users [post]
func (api *API) Post(w http.ResponseWriter, r *http.Request) {
	var body PostBodyParams
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	if err := api.validator.Struct(body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	existingUser, err := user.GetUserByEmail(body.Email, api.db)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			response.GenericServerError(w, err)
			return
		}
	}

	if existingUser != nil {
		response.GenericBadRequestError(w, fmt.Errorf("User already exists"))
		return
	}

	hashedPassword, err := auth.HashPassword(body.Password)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	userData := models.User{
		ID:       uuid.New(),
		Name:     body.Name,
		Email:    body.Email,
		Password: hashedPassword,
	}

	var user models.User
	err = api.db.Get(&user, "INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4) RETURNING *", userData.ID, userData.Name, userData.Email, userData.Password)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	user.Password = "redacted"

	api.log.Info.Printf("User %s has been created", user.ID)

	response.HTTPResponse(w, user)
}

// @Summary		Update User
// @Description	Update authenticated User with the parameters sent with the request based on the JWT
// @Tags			User
// @Accept	  json
// @Produce		json
// @Param			Path	path		UserPathParams	true	"UserPathParams"
// @Param			Body	body		PutBodyParams	true	"PutBodyParams"
// @Success		200		{object}	response.Response{data=models.User}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Security		BearerAuth
// @Router			/users/{user_id} [put]
func (api *API) Put(w http.ResponseWriter, r *http.Request) {
	path := UserPathParams{
		UserID: r.PathValue("user_id"),
	}

	if err := api.validator.Struct(path); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	var body PutBodyParams
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	if err := api.validator.Struct(body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	existingUser := user.GetUserFromJWT(r, w, api.db)
	if existingUser == nil {
		return
	}

	if existingUser.ID.String() != path.UserID {
		response.GenericUnauthenticatedError(w)
		return
	}

	userData := models.User{
		ID:    existingUser.ID,
		Name:  body.Name,
		Email: body.Email,
	}

	emailInUse, err := user.IsEmailInUse(userData.Email, userData.ID.String(), api.db)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	if emailInUse {
		response.GenericBadRequestError(w, fmt.Errorf("Email already in use"))
		return
	}

	var user models.User
	err = api.db.Get(&user, "UPDATE users SET name=$1, email=$2, updated_at=$3 WHERE id=$4 AND active=true RETURNING *", userData.Name, userData.Email, time.Now(), userData.ID.String())
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	user.Password = "redacted"

	api.log.Info.Printf("User %s has been updated", user.ID)

	response.HTTPResponse(w, user)
}

// @Summary		Delete User
// @Description	Delete authenticated User based on the JWT
// @Tags			User
// @Produce		json
// @Param			Path	path		UserPathParams	true	"UserPathParams"
// @Success		200		{object}	response.Response{data=string}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Security		BearerAuth
// @Router			/users/{user_id} [delete]
func (api *API) Delete(w http.ResponseWriter, r *http.Request) {
	path := UserPathParams{
		UserID: r.PathValue("user_id"),
	}

	if err := api.validator.Struct(path); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	user := user.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	if user.ID.String() != path.UserID {
		response.GenericUnauthenticatedError(w)
		return
	}

	res, err := api.db.Exec("UPDATE users SET active=false, deleted_at=$1 WHERE id=$2", time.Now(), user.ID)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	if count == 0 {
		response.HTTPError(w, http.StatusBadRequest, "User does not exist", response.StatusFail)
		return
	}

	api.log.Info.Printf("User %s has been deleted", user.ID)

	response.HTTPResponse(w, "User has been successfully deleted")
}
