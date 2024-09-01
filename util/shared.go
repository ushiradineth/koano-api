package util

import (
	"encoding/json"
	"net/http"

	validatorUtil "github.com/ushiradineth/cron-be/util/validator"
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

func GenericServerError(w http.ResponseWriter, err error) {
	HTTPError(w, http.StatusInternalServerError, err.Error(), StatusError)
}

func GenericValidationError(w http.ResponseWriter, err error) {
	HTTPError(w, http.StatusBadRequest, validatorUtil.ValidationError(err), StatusFail)
}
