package event

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func GetEventHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("event_id")

	event, err := GetEvent(id)
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

func PostEventHandler(w http.ResponseWriter, r *http.Request) {
	user_id := r.FormValue("user_id")
	title := r.FormValue("title")
	start := r.FormValue("start")
	end := r.FormValue("end")
	tz := r.FormValue("tz")
	repeated := r.FormValue("repeated")

	event, err := DoesEventExist("", start, end, user_id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get event data: %v", err), http.StatusInternalServerError)
		return
	}

	if event {
		http.Error(w, fmt.Sprintf("Event already exists"), http.StatusBadRequest)
		return
	}

	parsedStart, err := time.Parse(time.RFC3339, start)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}

	parsedEnd, err := time.Parse(time.RFC3339, end)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}

	_, err = DB.Exec("INSERT INTO event (id, title, start, end, user_id, tz, repeated) VALUES (?, ?, ?, ?, ?, ?, ?)", uuid.New(), title, parsedStart, parsedEnd, user_id, tz, repeated)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert event data: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func PutEventHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("event_id")
	title := r.FormValue("title")
	start := r.FormValue("start")
	end := r.FormValue("end")
	user_id := r.FormValue("user_id")
	tz := r.FormValue("tz")
	repeated := r.FormValue("repeated")

	event, err := DoesEventExist(id, start, end, user_id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get event data: %v", err), http.StatusInternalServerError)
		return
	}

	if !event {
		http.Error(w, fmt.Sprintf("Event does not exist"), http.StatusBadRequest)
		return
	}

	parsedStart, err := time.Parse(time.RFC3339, start)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}

	parsedEnd, err := time.Parse(time.RFC3339, end)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}

	_, err = DB.Exec("UPDATE event SET title=(?), start=(?), end=(?), tz=(?), repeated=(?) WHERE id=(?)", title, parsedStart, parsedEnd, tz, repeated, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert event data: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteEventHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("event_id")

	res, err := DB.Exec("DELETE FROM event WHERE id=(?)", id)
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

func GetUserEventsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("user_id")

	events := []Event{}
	err := DB.Select(&events, "SELECT * FROM event WHERE user_id=(?)", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user event data: %v", err), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(events)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal events data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	w.WriteHeader(http.StatusOK)
}
