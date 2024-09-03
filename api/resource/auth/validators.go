package auth

type AuthenticateQueryParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=20,hasLowercase,hasUppercase,hasDigit,hasSpecialCharacter"`
}

type RefreshTokenQueryParams struct {
	RefreshToken string `json:"refresh_token" validate:"required,jwt"`
}

type PutPasswordQueryParams struct {
	Password string `json:"password" validate:"required,min=8,max=20,hasLowercase,hasUppercase,hasDigit,hasSpecialCharacter"`
}
