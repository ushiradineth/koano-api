package auth

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
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

type AuthenticateResponse struct {
	models.User
	AccessToken  string
	RefreshToken string
}

type RefreshTokenResponse struct {
	AccessToken string
}

// @Summary		Authenticate User
// @Description	Authenticate User with the parameters sent with the request
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			Query	query		AuthenticateQueryParams	true	"AuthenticateQueryParams"
// @Success		200		{object}	response.Response{data=AuthenticateResponse}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Router			/auth/login [post]
func (api *API) Authenticate(w http.ResponseWriter, r *http.Request) {
	query := AuthenticateQueryParams{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := api.validator.Struct(query); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	user := user.GetUser(w, query.Email, api.db)
	if user == nil {
		return
	}

	valid := auth.CheckPasswordHash(query.Password, user.Password)

	if !valid {
		response.HTTPError(w, http.StatusUnauthorized, "Invalid Credentials", response.StatusFail)
		return
	}

	accessToken, err := auth.NewAccessToken(user.ID, user.Name, user.Email)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	refreshTokenClaim := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
	}

	refreshToken, err := auth.NewRefreshToken(refreshTokenClaim)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	user.Password = "redacted"

	authenticateResponse := AuthenticateResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	response.HTTPResponse(w, authenticateResponse)
}

// @Summary		Refresh Access Token
// @Description	Refresh Access Token with the parameters sent with the request based on the request based on the JWT
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			Query	query		RefreshTokenQueryParams	true	"RefreshTokenQueryParams"
// @Success		200		{object}	response.Response{data=RefreshTokenResponse}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Security		BearerAuth
// @Router			/auth/refresh [post]
func (api *API) RefreshToken(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetJWT(r)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	query := RefreshTokenQueryParams{
		RefreshToken: r.FormValue("refresh_token"),
	}

	if err := api.validator.Struct(query); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	accessTokenClaim, err := auth.ParseExpiredAccessToken(accessToken)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	user := user.GetUser(w, accessTokenClaim.Email, api.db)
	if user == nil {
		return
	}

	_, errr := auth.ParseRefreshToken(query.RefreshToken)
	if errr != nil {
		response.GenericServerError(w, err)
		return
	}

	newAccessToken, err := auth.NewAccessToken(user.ID, user.Name, user.Email)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	refreshTokenResponse := RefreshTokenResponse{
		AccessToken: newAccessToken,
	}

	response.HTTPResponse(w, refreshTokenResponse)
}

// @Summary		Update User Password
// @Description	Update authenticated user's Password with the parameters sent with the request based on the JWT
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			Query	query		PutPasswordQueryParams	true	"PutPasswordQueryParams"
// @Success		200		{object}	response.Response{data=string}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Security		BearerAuth
// @Router			/auth/reset-password [put]
func (api *API) PutPassword(w http.ResponseWriter, r *http.Request) {
	query := PutPasswordQueryParams{
		Password: r.FormValue("password"),
	}

	if err := api.validator.Struct(query); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	user := user.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	password, err := auth.HashPassword(query.Password)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	_, err = api.db.Exec("UPDATE users SET password=$1 WHERE id=$2", password, user.ID)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	response.HTTPResponse(w, "Password has being updated")
}
