package response

import (
	"net/http"

	validatorUtil "github.com/ushiradineth/cron-be/util/validator"
)

func GenericServerError(w http.ResponseWriter, err error) {
	HTTPError(w, http.StatusInternalServerError, err.Error(), StatusError)
}

func GenericValidationError(w http.ResponseWriter, err error) {
	HTTPError(w, http.StatusBadRequest, validatorUtil.ValidationError(err), StatusFail)
}

func GenericBadRequestError(w http.ResponseWriter, err error) {
	HTTPError(w, http.StatusBadRequest, err.Error(), StatusFail)
}

func GenericUnauthenticatedError(w http.ResponseWriter) {
	HTTPError(w, http.StatusUnauthorized, "Unauthorized", StatusFail)
}
