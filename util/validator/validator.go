package validator

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

func New() *validator.Validate {
	validate := validator.New()

	// Using the names which have been specified for JSON representations of structs, rather than normal Go field names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	err := validate.RegisterValidation("hasLowercase", hasLowercase)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = validate.RegisterValidation("hasUppercase", hasUppercase)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = validate.RegisterValidation("hasDigit", hasDigit)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = validate.RegisterValidation("hasSpecialCharacter", hasSpecialCharacter)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return validate
}

func ValidationError(err error) []string {
	if fieldErrors, ok := err.(validator.ValidationErrors); ok {
		resp := make([]string, len(fieldErrors))

		for i, err := range fieldErrors {
			switch err.Tag() {
			case "required":
				resp[i] = fmt.Sprintf("%s field is required", err.Field())
			case "min":
				resp[i] = fmt.Sprintf("%s must be at least %s characters length", err.Field(), err.Param())
			case "max":
				resp[i] = fmt.Sprintf("%s can't be more than %s characters length", err.Field(), err.Param())
			case "email":
				resp[i] = fmt.Sprintf("%s must be a valid email", err.Field())
			case "jwt":
				resp[i] = fmt.Sprintf("%s must be a JWT token", err.Field())
			case "uuid":
				resp[i] = fmt.Sprintf("%s must be a valid UUID", err.Field())
			case "timezone":
				resp[i] = fmt.Sprintf("%s must be a valid Timezone", err.Field())
			case "datetime":
				resp[i] = fmt.Sprintf("%s must follow `%s` format", err.Field(), err.Param())
			case "hasLowercase":
				resp[i] = fmt.Sprintf("%s must contain at least one lowercase character", err.Field())
			case "hasUppercase":
				resp[i] = fmt.Sprintf("%s must contain at least one uppercase character", err.Field())
			case "hasDigit":
				resp[i] = fmt.Sprintf("%s must contain at least one digit", err.Field())
			case "hasSpecialCharacter":
				resp[i] = fmt.Sprintf("%s must contain at least one special character", err.Field())
			default:
				resp[i] = fmt.Sprintf("something is wrong with %s; %s", err.Field(), err.Tag())
			}
		}

		return resp
	}
	return nil
}

func hasUppercase(fl validator.FieldLevel) bool {
	for _, char := range fl.Field().String() {
		if unicode.IsUpper(char) {
			return true
		}
	}
	return false
}

func hasLowercase(fl validator.FieldLevel) bool {
	for _, char := range fl.Field().String() {
		if unicode.IsLower(char) {
			return true
		}
	}
	return false
}

func hasDigit(fl validator.FieldLevel) bool {
	for _, char := range fl.Field().String() {
		if unicode.IsDigit(char) {
			return true
		}
	}
	return false
}

func hasSpecialCharacter(fl validator.FieldLevel) bool {
	for _, char := range fl.Field().String() {
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			return true
		}
	}
	return false
}
