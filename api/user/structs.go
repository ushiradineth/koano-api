package user

import "github.com/ushiradineth/cron-be/models"

type AuthenticateResponse struct {
	models.User
	AccessToken  string
	RefreshToken string
}

type RefreshTokenResponse struct {
	AccessToken string
}

