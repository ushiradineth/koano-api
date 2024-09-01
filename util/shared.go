package util

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

const (
	StatusSuccess = "success" // 2XX
	StatusFail    = "fail"    // 4XX
	StatusError   = "error"   // 5XX
)

func HTTPError(w http.ResponseWriter, code int, message interface{}, status string) {
	var error Error

	switch status {
	case StatusFail, StatusError:
		error = Error{
			Code:   code,
			Status: status,
			Error:  message,
		}
	default:
		error = Error{
			Code:   http.StatusInternalServerError,
			Status: StatusError,
			Error:  "Failed to generate the error",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(error.Code)

	err := json.NewEncoder(w).Encode(error)
	if err != nil {
		http.Error(w, "Failed to generate the error", http.StatusInternalServerError)
		return
	}
}

func HTTPResponse(w http.ResponseWriter, data interface{}) {
	response := Response{
		Code:   http.StatusOK,
		Status: StatusSuccess,
		Data:   data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Code)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to generate the response", http.StatusInternalServerError)
		return
	}
}

func ValidationError(err error) []string {
	if fieldErrors, ok := err.(validator.ValidationErrors); ok {
		resp := make([]string, len(fieldErrors))

		for i, err := range fieldErrors {
			switch err.Tag() {
			case "required":
				resp[i] = fmt.Sprintf("%s: field is required", err.Field())
			case "oneof":
				resp[i] = fmt.Sprintf("%s: field can only be: %s", err.Field(), err.ActualTag())
			case "min":
				resp[i] = fmt.Sprintf("%s must be at least %s characters length", err.Field(), err.ActualTag())
			case "max":
				resp[i] = fmt.Sprintf("%s can't be more that %s characters length", err.Field(), err.ActualTag())
			case "email":
				resp[i] = fmt.Sprintf("%s must be a valid email", err.Field())
			case "jwt":
				resp[i] = fmt.Sprintf("%s must be a JWT token", err.Field())
			case "uuid":
				resp[i] = fmt.Sprintf("%s must be a valid UUID", err.Field())
			case "timezone":
				resp[i] = fmt.Sprintf("%s must be a valid Timezone", err.Field())
			case "lowercase":
				resp[i] = fmt.Sprintf("%s must contain at least one lowercase character", err.Field())
			case "uppercase":
				resp[i] = fmt.Sprintf("%s must contain at least one uppercase character", err.Field())
			case "digitrequired":
				resp[i] = fmt.Sprintf("%s must contain at least one digit", err.Field())
			case "specialsymbol":
				resp[i] = fmt.Sprintf("%s must contain at least one special symbol", err.Field())
			case "datetime":
				resp[i] = fmt.Sprintf("%s must follow `%s` format", err.Field(), err.Param())
			default:
				resp[i] = fmt.Sprintf("something wrong on %s; %s", err.Field(), err.Tag())
			}
		}

		return resp
	}
	return nil
}

func GenericServerError(w http.ResponseWriter, err error) {
	HTTPError(w, http.StatusInternalServerError, err.Error(), StatusError)
}

func GenericValidationError(w http.ResponseWriter, err error) {
	HTTPError(w, http.StatusBadRequest, ValidationError(err), StatusFail)
}
