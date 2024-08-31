package user

type PostQueryParams struct {
	Name     string `form:"name" validate:"required"`
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required,min=8,max=20"`
}

type PutQueryParams struct {
	Name  string `form:"name" validate:"required"`
	Email string `form:"email" validate:"required,email"`
}

type PutPasswordQueryParams struct {
	Password string `form:"password" validate:"required,password"`
}

type AuthenticateQueryParams struct {
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required,min=8"`
}

type RefreshTokenQueryParams struct {
	RefreshToken string `form:"refresh_token" validate:"required,jwt"`
}
