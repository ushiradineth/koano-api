package health

import (
	"net/http"

	"github.com/ushiradineth/cron-be/util/response"
)

func Health(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte("healthy"))
	if err != nil {
		response.GenericServerError(w, err)
		return
	}
}
