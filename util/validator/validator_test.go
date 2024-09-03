package validator_test

import (
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/ushiradineth/cron-be/util/validator"
)

type testCase struct {
	name     string
	input    interface{}
	expected string
}

var errorTests = []*testCase{
	{
		name: `required`,
		input: struct {
			Title string `json:"title" validate:"required"`
		}{},
		expected: "title field is required",
	},
	{
		name: `min`,
		input: struct {
			ID string `json:"id" validate:"min=8"`
		}{ID: "1234567"},
		expected: "id must be at least 8 characters length",
	},
	{
		name: `max`,
		input: struct {
			ID string `json:"id" validate:"max=20"`
		}{ID: "123456789abcdefghijkl"},
		expected: "id can't be more than 20 characters length",
	},
	{
		name: `email`,
		input: struct {
			Email string `json:"email" validate:"email"`
		}{Email: "not_a_email"},
		expected: "email must be a valid email",
	},
	{
		name: `jwt`,
		input: struct {
			JWT string `json:"jwt" validate:"jwt"`
		}{JWT: "not_a_jwt"},
		expected: "jwt must be a JWT token",
	},
	{
		name: `uuid`,
		input: struct {
			ID string `json:"id" validate:"uuid"`
		}{ID: "not_a_id"},
		expected: "id must be a valid UUID",
	},
	{
		name: `timezone`,
		input: struct {
			Timezone string `json:"timezone" validate:"timezone"`
		}{Timezone: "not_a_tz"},
		expected: "timezone must be a valid Timezone",
	},
	{
		name: `datetime`,
		input: struct {
			Date string `json:"date" validate:"datetime=2006-01-02"`
		}{Date: "2006/01/02"},
		expected: "date must follow `2006-01-02` format",
	},
	{
		name: `hasLowercase`,
		input: struct {
			Name string `json:"name" validate:"hasLowercase"`
		}{Name: "UPPERCASE"},
		expected: "name must contain at least one lowercase character",
	},
	{
		name: `hasUppercase`,
		input: struct {
			Name string `json:"name" validate:"hasUppercase"`
		}{Name: "lowercase"},
		expected: "name must contain at least one uppercase character",
	},
	{
		name: `hasDigit`,
		input: struct {
			Name string `json:"name" validate:"hasDigit"`
		}{Name: "not_numbers"},
		expected: "name must contain at least one digit",
	},
	{
		name: `hasSpecialCharacter`,
		input: struct {
			Name string `json:"name" validate:"hasSpecialCharacter"`
		}{Name: "notsymbols"},
		expected: "name must contain at least one special character",
	},
	{
		name: `oneof`,
		input: struct {
			Repeated string `json:"repeated" validate:"oneof=never daily weekly monthly yearly"`
		}{Repeated: "nothing"},
		expected: "repeated field can only be one of the following `never daily weekly monthly yearly`",
	},
	{
		name: `default`,
		input: struct {
			URL string `json:"url" validate:"url"`
		}{URL: "image.png"},
		expected: "something is wrong with url; url",
	},
}

var successTests = []*testCase{
	{
		name: `required`,
		input: struct {
			Title string `json:"title" validate:"required"`
		}{Title: "title"},
		expected: "",
	},
	{
		name: `min`,
		input: struct {
			ID string `json:"id" validate:"min=8"`
		}{ID: "123456789"},
		expected: "",
	},
	{
		name: `max`,
		input: struct {
			ID string `json:"id" validate:"max=20"`
		}{ID: "123456789"},
		expected: "",
	},
	{
		name: `email`,
		input: struct {
			Email string `json:"email" validate:"email"`
		}{Email: "ushiradineth@gmail.com"},
		expected: "",
	},
	{
		name: `jwt`,
		input: struct {
			JWT string `json:"jwt" validate:"jwt"`
		}{JWT: func() string {
			token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "123456789", "name": "Ushira Dineth"}).SignedString([]byte("secret"))
			return token
		}()},
		expected: "",
	},
	{
		name: `uuid`,
		input: struct {
			ID string `json:"id" validate:"uuid"`
		}{ID: uuid.New().String()},
		expected: "",
	},
	{
		name: `timezone`,
		input: struct {
			Timezone string `json:"timezone" validate:"timezone"`
		}{Timezone: "Asia/Colombo"},
		expected: "",
	},
	{
		name: `datetime`,
		input: struct {
			Date string `json:"date" validate:"datetime=2006-01-02"`
		}{Date: "2006-01-02"},
		expected: "",
	},
	{
		name: `hasLowercase`,
		input: struct {
			Name string `json:"name" validate:"hasLowercase"`
		}{Name: "lowercase"},
		expected: "",
	},
	{
		name: `hasUppercase`,
		input: struct {
			Name string `json:"name" validate:"hasUppercase"`
		}{Name: "UPPERCASER"},
		expected: "",
	},
	{
		name: `hasDigit`,
		input: struct {
			Name string `json:"name" validate:"hasDigit"`
		}{Name: "1234"},
		expected: "",
	},
	{
		name: `hasSpecialCharacter`,
		input: struct {
			Name string `json:"name" validate:"hasSpecialCharacter"`
		}{Name: "!@#$%^"},
		expected: "",
	},
	{
		name: `default`,
		input: struct {
			URL string `json:"url" validate:"url"`
		}{URL: "https://ushira.com"},
		expected: "",
	},
}

func TestToErrorResponse(t *testing.T) {
	vr := validator.New()

	for _, tc := range errorTests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := vr.Struct(tc.input)
			errResp := validator.ValidationError(err)
			if errResp == nil || len(errResp) != 1 {
				t.Fatalf(`Expected:"%v", Got:"%v"`, tc.expected, errResp)
			} else if errResp[0] != tc.expected {
				t.Fatalf(`Expected:"%v", Got:"%v"`, tc.expected, errResp[0])
			}
		})
	}
}

func TestToSuccessResponse(t *testing.T) {
	vr := validator.New()

	for _, tc := range successTests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := vr.Struct(tc.input)
			errResp := validator.ValidationError(err)
			if errResp != nil || len(errResp) != 0 {
				t.Fatalf(`Expected:"%v", Got:"%v"`, tc.expected, errResp)
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	validate := validator.New()

	if validate == nil {
		t.Fatalf("Expected validator instance, got nil")
	}
}
