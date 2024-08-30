package user

type PostForm struct {
	Name     string `form:"name" validate:"required"`
	Email    string `form:"email" validate:"required,email"`
  Password string `form:"password" validate:"required,min=8,max=20"`
}

type PutForm struct {
	Name  string `form:"name" validate:"required"`
	Email string `form:"email" validate:"required,email"`
}

type PutPasswordForm struct {
  Password string `form:"password" validate:"required,password"`
}

type AuthenticateForm struct {
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required,min=8"`
}

type RefreshTokenForm struct {
	RefreshToken string `form:"refresh_token" validate:"required,jwt"`
}

