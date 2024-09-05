package response

import (
	"encoding/json"
	"net/http"
)

const (
	StatusSuccess = "success" // 2XX
	StatusFail    = "fail"    // 4XX
	StatusError   = "error"   // 5XX
)

type Error struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Error  interface{} `json:"error"`
}

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

type Response struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
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
