package auth

type AuthenticateQueryParams struct {
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required,min=8,max=20,hasLowercase,hasUppercase,hasDigit,hasSpecialCharacter"`
}

type RefreshTokenQueryParams struct {
	RefreshToken string `form:"refresh_token" validate:"required,jwt"`
}

type PutPasswordQueryParams struct {
	Password string `form:"password" validate:"required,min=8,max=20,hasLowercase,hasUppercase,hasDigit,hasSpecialCharacter"`
}
