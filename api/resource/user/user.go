package user

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ushiradineth/cron-be/models"
	"github.com/ushiradineth/cron-be/util"
)

type API struct {
	db        *sqlx.DB
	validator *validator.Validate
}

func New(db *sqlx.DB, validator *validator.Validate) *API {
	return &API{
		db:        db,
		validator: validator,
	}
}

// @Summary		Get User
// @Description	Get authenticated user based on the JWT sent with the request
// @Tags			User
// @Produce		json
// @Success		200	{object}	util.Response{data=models.User}
// @Failure		400	{object}	util.Error
// @Failure		401	{object}	util.Error
// @Failure		500	{object}	util.Error
// @Security		BearerAuth
// @Router			/user [get]
func (api *API) Get(w http.ResponseWriter, r *http.Request) {
	user := util.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	user.Password = "redacted"

	util.HTTPResponse(w, user)
}

// @Summary		Create User
// @Description	Create User with the parameters sent with the request
// @Tags			User
// @Accept			json
// @Produce		json
// @Param			Query	query		PostQueryParams	true	"PostQueryParams"
// @Success		200		{object}	util.Response{data=models.User}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Router			/user [post]
func (api *API) Post(w http.ResponseWriter, r *http.Request) {
	query := PostQueryParams{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := api.validator.Struct(query); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	userExists, _, err := util.DoesUserExist("", query.Email, api.db)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	if userExists {
		util.HTTPError(w, http.StatusBadRequest, "User already exists", util.StatusFail)
		return
	}

	hashedPassword, err := util.HashPassword(query.Password)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	user := models.User{
		ID:        uuid.New(),
		Name:      query.Name,
		Email:     query.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}

	_, err = api.db.Exec("INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4)", user.ID, user.Name, user.Email, user.Password)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	user.Password = "redacted"

	util.HTTPResponse(w, user)
}

// @Summary		Update User
// @Description	Update authenticated User with the parameters sent with the request based on the JWT
// @Tags			User
// @Accept			json
// @Produce		json
// @Param			Query	query		PutQueryParams	true	"PutQueryParams"
// @Success		200		{object}	util.Response{data=models.User}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Security		BearerAuth
// @Router			/user [put]
func (api *API) Put(w http.ResponseWriter, r *http.Request) {
	query := PutQueryParams{
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
	}

	if err := api.validator.Struct(query); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	existingUser := util.GetUserFromJWT(r, w, api.db)
	if existingUser == nil {
		return
	}

	user := models.User{
		ID:        existingUser.ID,
		Name:      query.Name,
		Email:     query.Email,
		CreatedAt: existingUser.CreatedAt,
		Password:  "redacted",
	}

	_, count, err := util.DoesUserExist(user.ID.String(), user.Email, api.db)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	if count > 1 {
		util.HTTPError(w, http.StatusBadRequest, "Email already in use", util.StatusFail)
		return
	}

	_, err = api.db.Exec("UPDATE users SET name=$1, email=$2 WHERE id=$3", user.Name, user.Email, user.ID.String())
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	util.HTTPResponse(w, user)
}

// @Summary		Delete User
// @Description	Delete authenticated User based on the JWT
// @Tags			User
// @Accept			json
// @Produce		json
// @Success		200	{object}	util.Response{data=string}
// @Failure		400	{object}	util.Error
// @Failure		401	{object}	util.Error
// @Failure		500	{object}	util.Error
// @Security		BearerAuth
// @Router			/user [delete]
func (api *API) Delete(w http.ResponseWriter, r *http.Request) {
	user := util.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	res, err := api.db.Exec("DELETE FROM users WHERE id=$1", user.ID)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	if count == 0 {
		util.HTTPError(w, http.StatusBadRequest, "User does not exist", util.StatusError)
		return
	}

	util.HTTPResponse(w, "User has been successfully deleted")
}
