package event

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ushiradineth/cron-be/models"
	"github.com/ushiradineth/cron-be/util"
)

func GetEventHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, err := util.GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), http.StatusInternalServerError)
		return
	}

	id := r.PathValue("event_id")

	event, err := util.GetEvent(id, user.ID.String(), db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get event data: %v", err), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(event)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal event data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	w.WriteHeader(http.StatusOK)
}

func PostEventHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, err := util.GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	start_time := r.FormValue("start_time")
	end_time := r.FormValue("end_time")
	timezone := r.FormValue("timezone")
	repeated := r.FormValue("repeated")

	event := util.DoesEventExist("", start_time, end_time, user.ID.String(), db)

	if event {
		http.Error(w, fmt.Sprintf("Event already exists"), http.StatusBadRequest)
		return
	}

	parsed_start, err := time.Parse(time.RFC3339, start_time)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}

	parsed_end, err := time.Parse(time.RFC3339, end_time)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}

	_, err = db.Exec("INSERT INTO events (id, title, start_time, end_time, user_id, timezone, repeated) VALUES ($1, $2, $3, $4, $5, $6, $7)", uuid.New(), title, parsed_start, parsed_end, user.ID, timezone, repeated)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert event data: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func PutEventHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, err := util.GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), http.StatusInternalServerError)
		return
	}

	id := r.PathValue("event_id")
	title := r.FormValue("title")
	start_time := r.FormValue("start_time")
	end_time := r.FormValue("end_time")
	timezone := r.FormValue("timezone")
	repeated := r.FormValue("repeated")

	event := util.DoesEventExist(id, start_time, end_time, user.ID.String(), db)

	if !event {
		http.Error(w, fmt.Sprintf("Event does not exist"), http.StatusBadRequest)
		return
	}

	parsed_start, err := time.Parse(time.RFC3339, start_time)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}

	parsed_end, err := time.Parse(time.RFC3339, end_time)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}

	_, err = db.Exec("UPDATE events SET title=$1, start_time=$2, end_time=$3, timezone=$4, repeated=$5 WHERE id=$6 AND user_id=$7", title, parsed_start, parsed_end, timezone, repeated, id, user.ID.String())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert event data: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteEventHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, err := util.GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), http.StatusInternalServerError)
		return
	}

	id := r.PathValue("event_id")

	res, err := db.Exec("DELETE FROM events WHERE id=$1 AND user_id=$2", id, user.ID.String())
	if err != nil {
		http.Error(w, fmt.Sprintf("Event does not exist"), http.StatusBadRequest)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("Event does not exist"), http.StatusBadRequest)
		return
	}

	if count == 0 {
		http.Error(w, fmt.Sprintf("Event does not exist"), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetUserEventsHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, err := util.GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), http.StatusInternalServerError)
		return
	}

	events := []models.Event{}
	db.Select(&events, "SELECT * FROM events WHERE user_id=$1", user.ID)

	response, err := json.Marshal(events)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal events data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
