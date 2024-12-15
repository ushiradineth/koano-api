package auth

type AuthenticateBodyParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=20,hasLowercase,hasUppercase,hasDigit,hasSpecialCharacter"`
}

type RefreshTokenBodyParams struct {
	RefreshToken string `json:"refresh_token" validate:"required,jwt"`
}

type PutPasswordBodyParams struct {
	Password string `json:"password" validate:"required,min=8,max=20,hasLowercase,hasUppercase,hasDigit,hasSpecialCharacter"`
}
