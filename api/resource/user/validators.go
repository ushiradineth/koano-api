package user

type UserPathParams struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

type PostQueryParams struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=20,hasLowercase,hasUppercase,hasDigit,hasSpecialCharacter"`
}

type PutQueryParams struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}
