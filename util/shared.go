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

func HTTPError(w http.ResponseWriter, code int, message string, status string) {
	var error Error

	switch status {
	case StatusFail, StatusError:
		error = Error{
			Code:    code,
			Status:  status,
			Message: message,
		}
	default:
		error = Error{
			Code:    http.StatusInternalServerError,
			Status:  StatusError,
			Message: "Failed to generate the error",
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
	var response Response

	response = Response{
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

func ValidationError(errors validator.ValidationErrors) string {
	err := errors[0]

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s: field is required", err.Field())
	case "oneof":
		return fmt.Sprintf("%s: field can only be: %s", err.Field(), err.ActualTag())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters length", err.Field(), err.ActualTag())
	case "max":
		return fmt.Sprintf("%s can't be more that %s characters length", err.Field(), err.ActualTag())
	case "email":
		return fmt.Sprintf("%s must be a valid email", err.Field())
	case "jwt":
		return fmt.Sprintf("%s must be a JWT token", err.Field())
  case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", err.Field())
  case "timezone":
		return fmt.Sprintf("%s must be a valid Timezone", err.Field())
	case "lowercase":
		return fmt.Sprintf("%s must contain at least one lowercase character", err.Field())
	case "uppercase":
		return fmt.Sprintf("%s must contain at least one uppercase character", err.Field())
	case "digitrequired":
		return fmt.Sprintf("%s must contain at least one digit", err.Field())
	case "specialsymbol":
		return fmt.Sprintf("%s must contain at least one special symbol", err.Field())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}
