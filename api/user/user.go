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
// @Description	Get authenticated user based on the JWT sent with the request
// @Tags			User
// @Produce		json
// @Success		200	{object}	util.Response{data=models.User}
// @Failure		400	{object}	util.Error
// @Failure		401	{object}	util.Error
// @Failure		500	{object}	util.Error
// @Security		BearerAuth
// @Router			/user [get]
func Get(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user := util.GetUserFromJWT(r, w, db)

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
func Post(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	query := PostQueryParams{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := validate.Struct(query); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	userExists, _, err := util.DoesUserExist("", query.Email, db)

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

	_, err = db.Exec("INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4)", user.ID, user.Name, user.Email, user.Password)
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
func Put(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	query := PutQueryParams{
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
	}

	if err := validate.Struct(query); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	existingUser := util.GetUserFromJWT(r, w, db)

	user := models.User{
		ID:        existingUser.ID,
		Name:      query.Name,
		Email:     query.Email,
		CreatedAt: existingUser.CreatedAt,
		Password:  "redacted",
	}

	_, count, err := util.DoesUserExist(user.ID.String(), user.Email, db)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	if count > 1 {
		util.HTTPError(w, http.StatusBadRequest, "Email already in use", util.StatusFail)
		return
	}

	_, err = db.Exec("UPDATE users SET name=$1, email=$2 WHERE id=$3", user.Name, user.Email, user.ID.String())
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	util.HTTPResponse(w, user)
}

// @Summary		Update User Password
// @Description	Update authenticated User Password with the parameters sent with the request based on the JWT
// @Tags			User
// @Accept			json
// @Produce		json
// @Param			Query	query		PutPasswordQueryParams	true	"PutPasswordQueryParams"
// @Success		200		{object}	util.Response{data=string}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Security		BearerAuth
// @Router			/user/auth/password [put]
func PutPassword(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	query := PutPasswordQueryParams{
		Password: r.FormValue("password"),
	}

	if err := validate.Struct(query); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	user := util.GetUserFromJWT(r, w, db)

	password, err := util.HashPassword(query.Password)
	if err != nil {
		util.GenericServerError(w, err)
	}

	_, err = db.Exec("UPDATE users SET password=$1 WHERE id=$2", password, user.ID)
	if err != nil {
		util.GenericServerError(w, err)
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
// @Failure		500	{object}	util.Error
// @Security		BearerAuth
// @Router			/user [delete]
func Delete(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user := util.GetUserFromJWT(r, w, db)

	res, err := db.Exec("DELETE FROM users WHERE id=$1", user.ID)
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

// @Summary		Authenticate User
// @Description	Authenticated User with the parameters sent with the request
// @Tags			User
// @Accept			json
// @Produce		json
// @Param			Query	query		AuthenticateQueryParams	true	"AuthenticateQueryParams"
// @Success		200		{object}	util.Response{data=AuthenticateResponse}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Router			/user/auth [post]
func Authenticate(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	query := AuthenticateQueryParams{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := validate.Struct(query); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	user := util.GetUser(w, query.Email, db)

	valid := util.CheckPasswordHash(query.Password, user.Password)

	if !valid {
		util.HTTPError(w, http.StatusUnauthorized, "Invalid Credentials", util.StatusFail)
		return
	}

	accessToken, err := util.NewAccessToken(user.ID, user.Name, user.Email)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	refreshTokenClaim := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
	}

	refreshToken, err := util.NewRefreshToken(refreshTokenClaim)
	if err != nil {
		util.GenericServerError(w, err)
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
// @Param			Query	query		RefreshTokenQueryParams	true	"RefreshTokenQueryParams"
// @Success		200		{object}	util.Response{data=RefreshTokenResponse}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Security		BearerAuth
// @Router			/user/auth/refresh [post]
func RefreshToken(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	accessToken, err := util.GetJWT(r)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	query := RefreshTokenQueryParams{
		RefreshToken: r.FormValue("refresh_token"),
	}

	if err := validate.Struct(query); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	accessTokenClaim, err := util.ParseExpiredAccessToken(accessToken)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	user := util.GetUser(w, accessTokenClaim.Email, db)

	_, errr := util.ParseRefreshToken(query.RefreshToken)
	if errr != nil {
		util.GenericServerError(w, err)
		return
	}

	newAccessToken, err := util.NewAccessToken(user.ID, user.Name, user.Email)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	response := RefreshTokenResponse{
		AccessToken: newAccessToken,
	}

	util.HTTPResponse(w, response)
}
