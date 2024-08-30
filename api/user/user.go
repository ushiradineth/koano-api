package user

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ushiradineth/cron-be/models"
	"github.com/ushiradineth/cron-be/util"
)

var validate = validator.New()

// @Summary		Get User
//
// @Description	Get authenticated user based on the JWT sent with the request
//
// @Tags			User
//
// @Produce		json
//
// @Success		200	{object}	util.Response{data=models.User}
//
// @Failure		400	{object}	util.Error
// @Failure		401	{object}	util.Error
// @Failure		500	{object}	util.Error
// @Security		BearerAuth
// @Router			/user [get]
func Get(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, code, status, err := util.GetUserFromJWT(r, db)
	if err != nil {
		util.HTTPError(w, code, err.Error(), status)
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
// @Param			Form	query		PostForm	true "PostForm"
// @Success		200		{object}	util.Response{data=models.User}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Router			/user [post]
func Post(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	form := PostForm{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := validate.Struct(form); err != nil {
		util.HTTPError(w, http.StatusBadRequest, err.Error(), util.StatusFail)
		return
	}

	userExists, _, err := util.DoesUserExist("", form.Email, db)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	if userExists {
		util.HTTPError(w, http.StatusBadRequest, "User already exists", util.StatusFail)
		return
	}

	hashedPassword, err := util.HashPassword(form.Password)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	user := models.User{
		ID:        uuid.New(),
		Name:      form.Name,
		Email:     form.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}

	_, err = db.Exec("INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4)", user.ID, user.Name, user.Email, user.Password)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	user.Password = "redacted"

	util.HTTPResponse(w, user)
}

// @Summary		Update User
// @Description	Update authenticated User with the parameters sent with the request based on the JWT
// @Tags			User
//
// @Accept			json
// @Produce		json
// @Param			Form	query		PutForm	true "PutForm"
// @Success		200		{object}	util.Response{data=models.User}
// @Success		200		{object}	models.User
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Security		BearerAuth
// @Router			/user [put]
func Put(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	form := PutForm{
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
	}

	if err := validate.Struct(form); err != nil {
		util.HTTPError(w, http.StatusBadRequest, err.Error(), util.StatusFail)
		return
	}

	existingUser, code, status, err := util.GetUserFromJWT(r, db)
	if err != nil {
		util.HTTPError(w, code, err.Error(), status)
		return
	}

	user := models.User{
		ID:        existingUser.ID,
		Name:      form.Name,
		Email:     form.Email,
		CreatedAt: existingUser.CreatedAt,
		Password:  "redacted",
	}

	_, count, err := util.DoesUserExist(user.ID.String(), user.Email, db)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	if count > 1 {
		util.HTTPError(w, http.StatusBadRequest, "Email already in use", util.StatusFail)
		return
	}

	_, err = db.Exec("UPDATE users SET name=$1, email=$2 WHERE id=$3", user.Name, user.Email, user.ID.String())
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	util.HTTPResponse(w, user)
}

// @Summary		Update User Password
// @Description	Update authenticated User Password with the parameters sent with the request based on the JWT
// @Tags			User
// @Accept			json
// @Produce		json
// @Param			Form	query		PutPasswordForm	true "PutPasswordForm"
// @Success		200		{object}	util.Response{data=string}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Security		BearerAuth
// @Router			/user/auth/password [put]
func PutPassword(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	form := PutPasswordForm{
		Password: r.FormValue("password"),
	}

	if err := validate.Struct(form); err != nil {
		util.HTTPError(w, http.StatusBadRequest, err.Error(), util.StatusFail)
		return
	}

	user, code, status, err := util.GetUserFromJWT(r, db)
	if err != nil {
		util.HTTPError(w, code, err.Error(), status)
		return
	}

	password, err := util.HashPassword(form.Password)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
	}

	_, err = db.Exec("UPDATE users SET password=$1 WHERE id=$2", password, user.ID)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	util.HTTPResponse(w, "Password has being updated")
}

// @Summary		Delete User
// @Description	Delete authenticated User based on the JWT
// @Tags			User
// @Accept			json
// @Produce		json
// @Success		200	{object}	util.Response{data=string}
// @Failure		400	{object}	util.Error
// @Failure		401	{object}	util.Error
//
// @Failure		500	{object}	util.Error
//
// @Security		BearerAuth
//
// @Router			/user [delete]
func Delete(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, code, status, err := util.GetUserFromJWT(r, db)
	if err != nil {
		util.HTTPError(w, code, err.Error(), status)
		return
	}

	res, err := db.Exec("DELETE FROM users WHERE id=$1", user.ID)
	if err != nil {
		util.HTTPError(w, http.StatusBadRequest, err.Error(), util.StatusFail)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	if count == 0 {
		util.HTTPError(w, http.StatusBadRequest, "User does not exist", util.StatusError)
		return
	}

	util.HTTPResponse(w, "User has been successfully deleted")
}

// @Summary		Authenticate User
// @Description	Authenticated User with the parameters sent with the request
// @Tags			User
// @Accept			json
// @Produce		json
// @Param			Form	query		AuthenticateForm	true "AuthenticateForm"
// @Success		200		{object}	util.Response{data=AuthenticateResponse}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Router			/user/auth [post]
func Authenticate(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	form := AuthenticateForm{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := validate.Struct(form); err != nil {
		util.HTTPError(w, http.StatusBadRequest, err.Error(), util.StatusFail)
		return
	}

	user, err := util.GetUser(form.Email, db)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	valid := util.CheckPasswordHash(form.Password, user.Password)

	if !valid {
		util.HTTPError(w, http.StatusUnauthorized, "Invalid Credentials", util.StatusFail)
		return
	}

	accessToken, err := util.NewAccessToken(user.ID, user.Name, user.Email)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	refreshTokenClaim := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
	}

	refreshToken, err := util.NewRefreshToken(refreshTokenClaim)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	user.Password = "redacted"

	response := AuthenticateResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	util.HTTPResponse(w, response)
}

// @Summary		Refresh Access Token
// @Description	Refresh Access Token User with the parameters sent with the request based on the request based on the JWT
// @Tags			User
// @Accept			json
// @Produce		json
// @Param			Form	query		RefreshTokenForm	true "RefreshTokenForm"
// @Success		200		{object}	util.Response{data=RefreshTokenResponse}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Security		BearerAuth
// @Router			/user/auth/refresh [post]
func RefreshToken(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	accessToken, err := util.GetJWT(r)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	form := RefreshTokenForm{
		RefreshToken: r.FormValue("refresh_token"),
	}

	if err := validate.Struct(form); err != nil {
		util.HTTPError(w, http.StatusBadRequest, err.Error(), util.StatusFail)
		return
	}

	accessTokenClaim, err := util.ParseExpiredAccessToken(accessToken)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	_, errr := util.ParseRefreshToken(form.RefreshToken)
	if errr != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	newAccessToken, err := util.NewAccessToken(accessTokenClaim.Id, accessTokenClaim.Name, accessTokenClaim.Email)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	response := RefreshTokenResponse{
		AccessToken: newAccessToken,
	}

	util.HTTPResponse(w, response)
}
