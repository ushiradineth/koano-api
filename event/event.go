package event

import (
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func GetEventHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("event_id")
	fmt.Fprintf(w, "GET event with id=%v\n", id)
}

func PostEventHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "POST event\n")

	user_id := r.FormValue("user_id")
	title := r.FormValue("title")
	start := r.FormValue("start")
	end := r.FormValue("end")
	fmt.Fprintf(w, "UserID = %s\n", user_id)
	fmt.Fprintf(w, "Title = %s\n", title)
	fmt.Fprintf(w, "Start = %s\n", end)
	fmt.Fprintf(w, "End = %s\n", start)
}

func UpdateEventHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("eventid")
	fmt.Fprintf(w, "PUT event with id=%v\n", id)

	title := r.FormValue("title")
	start := r.FormValue("start")
	end := r.FormValue("end")
	fmt.Fprintf(w, "Title = %s\n", title)
	fmt.Fprintf(w, "Start = %s\n", end)
	fmt.Fprintf(w, "End = %s\n", start)
}

func GetUserEventsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("user_id")
	fmt.Fprintf(w, "GET events with userid=%v\n", id)
}
