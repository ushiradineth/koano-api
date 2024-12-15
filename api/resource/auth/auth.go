package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/ushiradineth/cron-be/models"
	"github.com/ushiradineth/cron-be/util/auth"
	logger "github.com/ushiradineth/cron-be/util/log"
	"github.com/ushiradineth/cron-be/util/response"
	"github.com/ushiradineth/cron-be/util/user"
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

type AuthenticateResponse struct {
	User         models.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int64       `json:"expires_in"`
	ExpiresAt    int64       `json:"expires_at"`
	RefreshToken string      `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

// @Summary		Authenticate User
// @Description	Authenticate User with the parameters sent with the request
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			Body	body		AuthenticateBodyParams	true	"AuthenticateBodyParams"
// @Success		200		{object}	response.Response{data=AuthenticateResponse}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Router			/auth/login [post]
func (api *API) Authenticate(w http.ResponseWriter, r *http.Request) {
	var body AuthenticateBodyParams
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	if err := api.validator.Struct(body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	user := user.GetUser(w, body.Email, api.db)
	if user == nil {
		return
	}

	valid := auth.CheckPasswordHash(body.Password, user.Password)

	if !valid {
		response.HTTPError(w, http.StatusUnauthorized, "Invalid Credentials", response.StatusFail)
		return
	}

	accessToken, expiresIn, expiresAt, err := auth.NewAccessToken(user.ID, user.Name, user.Email)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	refreshToken, err := auth.NewRefreshToken()
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	user.Password = "redacted"

	authenticateResponse := AuthenticateResponse{
		User:         *user,
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		ExpiresAt:    expiresAt,
		RefreshToken: refreshToken,
	}

	api.log.Info.Printf("User %s has been authenticated", user.ID)

	response.HTTPResponse(w, authenticateResponse)
}

// @Summary		Refresh Access Token
// @Description	Refresh Access Token with the parameters sent with the request based on the request based on the JWT
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			Body	body		RefreshTokenBodyParams	true	"RefreshTokenBodyParams"
// @Success		200		{object}	response.Response{data=RefreshTokenResponse}
// @Failure		400		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Security		BearerAuth
// @Router			/auth/refresh [post]
func (api *API) RefreshToken(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetJWT(r)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	var body RefreshTokenBodyParams
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	if err := api.validator.Struct(body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	accessTokenClaim := auth.ParseExpiredAccessToken(w, accessToken)
	if accessTokenClaim == nil {
		return
	}

	user := user.GetUser(w, accessTokenClaim.Email, api.db)
	if user == nil {
		return
	}

	refreshTokenClaim := auth.ParseRefreshToken(w, body.RefreshToken)
	if refreshTokenClaim == nil {
		return
	}

	newAccessToken, expiresIn, expiresAt, err := auth.NewAccessToken(user.ID, user.Name, user.Email)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	newRefreshToken, err := auth.NewRefreshToken()
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	refreshTokenResponse := RefreshTokenResponse{
		AccessToken:  newAccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		ExpiresAt:    expiresAt,
		RefreshToken: newRefreshToken,
	}

	api.log.Info.Printf("Access Token for user %s has been refreshed", user.ID)

	response.HTTPResponse(w, refreshTokenResponse)
}

// @Summary		Update User Password
// @Description	Update authenticated user's Password with the parameters sent with the request based on the JWT
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			Body	body		PutPasswordBodyParams	true	"PutPasswordBodyParams"
// @Success		200		{object}	response.Response{data=string}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Security		BearerAuth
// @Router			/auth/reset-password [put]
func (api *API) PutPassword(w http.ResponseWriter, r *http.Request) {
	var body PutPasswordBodyParams
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	if err := api.validator.Struct(body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	if err := api.validator.Struct(body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	user := user.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	password, err := auth.HashPassword(body.Password)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	_, err = api.db.Exec("UPDATE users SET password=$1 WHERE id=$2", password, user.ID)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	api.log.Info.Printf("User %s has updated their password", user.ID)

	response.HTTPResponse(w, "Password has being updated")
}
