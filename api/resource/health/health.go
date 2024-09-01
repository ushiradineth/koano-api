package health

import (
	"net/http"

	"github.com/ushiradineth/cron-be/util"
)

func Health(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte("healthy"))
	if err != nil {
		util.GenericServerError(w, err)
		return
	}
}
