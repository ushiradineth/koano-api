package user

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ushiradineth/cron-be/models"
	"github.com/ushiradineth/cron-be/util/auth"
	"github.com/ushiradineth/cron-be/util/response"
	"github.com/ushiradineth/cron-be/util/user"
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

//	@Summary		Get User
//	@Description	Get authenticated user based on the JWT sent with the request
//	@Tags			User
//	@Produce		json
//	@Param			Path	path		UserPathParams	true	"UserPathParams"
//	@Success		200		{object}	response.Response{data=models.User}
//	@Failure		400		{object}	response.Error
//	@Failure		401		{object}	response.Error
//	@Failure		500		{object}	response.Error
//	@Security		BearerAuth
//	@Router			/users/{user_id} [get]
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

	response.HTTPResponse(w, user)
}

//	@Summary		Create User
//	@Description	Create User with the parameters sent with the request
//	@Tags			User
//	@Accept			application/x-www-form-urlencoded
//	@Produce		json
//	@Param			Query	query		PostQueryParams	true	"PostQueryParams"
//	@Success		200		{object}	response.Response{data=models.User}
//	@Failure		400		{object}	response.Error
//	@Failure		401		{object}	response.Error
//	@Failure		500		{object}	response.Error
//	@Router			/users [post]
func (api *API) Post(w http.ResponseWriter, r *http.Request) {
	query := PostQueryParams{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := api.validator.Struct(query); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	userExists, _, err := user.DoesUserExist("", query.Email, api.db)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	if userExists {
		response.HTTPError(w, http.StatusBadRequest, "User already exists", response.StatusFail)
		return
	}

	hashedPassword, err := auth.HashPassword(query.Password)
	if err != nil {
		response.GenericServerError(w, err)
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
		response.GenericServerError(w, err)
		return
	}

	user.Password = "redacted"

	response.HTTPResponse(w, user)
}

//	@Summary		Update User
//	@Description	Update authenticated User with the parameters sent with the request based on the JWT
//	@Tags			User
//	@Accept			application/x-www-form-urlencoded
//	@Produce		json
//	@Param			Path	path		UserPathParams	true	"UserPathParams"
//	@Param			Query	query		PutQueryParams	true	"PutQueryParams"
//	@Success		200		{object}	response.Response{data=models.User}
//	@Failure		400		{object}	response.Error
//	@Failure		401		{object}	response.Error
//	@Failure		500		{object}	response.Error
//	@Security		BearerAuth
//	@Router			/users/{user_id} [put]
func (api *API) Put(w http.ResponseWriter, r *http.Request) {
	path := UserPathParams{
		UserID: r.PathValue("user_id"),
	}

	query := PutQueryParams{
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
	}

	if err := api.validator.Struct(path); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	if err := api.validator.Struct(query); err != nil {
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

	newUser := models.User{
		ID:        existingUser.ID,
		Name:      query.Name,
		Email:     query.Email,
		CreatedAt: existingUser.CreatedAt,
		Password:  "redacted",
	}

	_, count, err := user.DoesUserExist(newUser.ID.String(), newUser.Email, api.db)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	if count > 1 {
		response.HTTPError(w, http.StatusBadRequest, "Email already in use", response.StatusFail)
		return
	}

	_, err = api.db.Exec("UPDATE users SET name=$1, email=$2 WHERE id=$3", newUser.Name, newUser.Email, newUser.ID.String())
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	response.HTTPResponse(w, newUser)
}

//	@Summary		Delete User
//	@Description	Delete authenticated User based on the JWT
//	@Tags			User
//	@Accept			application/x-www-form-urlencoded
//	@Produce		json
//	@Param			Path	path		UserPathParams	true	"UserPathParams"
//	@Success		200		{object}	response.Response{data=string}
//	@Failure		400		{object}	response.Error
//	@Failure		401		{object}	response.Error
//	@Failure		500		{object}	response.Error
//	@Security		BearerAuth
//	@Router			/users/{user_id} [delete]
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

	res, err := api.db.Exec("DELETE FROM users WHERE id=$1", user.ID)
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

	response.HTTPResponse(w, "User has been successfully deleted")
}
