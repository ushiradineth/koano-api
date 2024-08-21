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

func Get(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, code, err := util.GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), code)
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

	w.WriteHeader(http.StatusOK)

	_, err = w.Write(response)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		fmt.Printf("Error writing response: %v\n", err)
		return
	}
}

func Post(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, code, err := util.GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), code)
		return
	}

	start_time := r.FormValue("start_time")
	end_time := r.FormValue("end_time")

	eventExists := util.DoesEventExist("", start_time, end_time, user.ID.String(), db)
	if eventExists {
		http.Error(w, "Event already exists", http.StatusBadRequest)
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

	event := models.Event{
		ID:       uuid.New(),
		Title:    r.FormValue("title"),
		Start:    parsed_start,
		End:      parsed_end,
		UserID:   user.ID,
		Timezone: r.FormValue("timezone"),
		Repeated: r.FormValue("repeated"),
	}

	_, err = db.Exec("INSERT INTO events (id, title, start_time, end_time, user_id, timezone, repeated) VALUES ($1, $2, $3, $4, $5, $6, $7)", event.ID, event.Title, event.Start, event.End, event.UserID, event.Timezone, event.Repeated)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert event data: %v", err), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(event)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal event data: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	_, err = w.Write(response)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		fmt.Printf("Error writing response: %v\n", err)
		return
	}
}

func Put(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, code, err := util.GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), code)
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
		http.Error(w, "Event does not exist", http.StatusBadRequest)
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

func Delete(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, code, err := util.GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), code)
		return
	}

	id := r.PathValue("event_id")

	res, err := db.Exec("DELETE FROM events WHERE id=$1 AND user_id=$2", id, user.ID.String())
	if err != nil {
		http.Error(w, "Event does not exist", http.StatusBadRequest)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Event does not exist", http.StatusBadRequest)
		return
	}

	if count == 0 {
		http.Error(w, "Event does not exist", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetUserEvents(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	startDay := r.PathValue("start_day")
	endDay := r.PathValue("end_day")

	startTime, err := time.Parse("2006-01-02", startDay)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse start day: %v", err), http.StatusBadRequest)
		return
	}
	endTime, err := time.Parse("2006-01-02", endDay)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse end day: %v", err), http.StatusBadRequest)
		return
	}

	user, code, err := util.GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), code)
		return
	}

	events := []models.Event{}

	err = db.Select(&events, "SELECT * FROM events WHERE user_id=$1 AND start_time >= $2 AND start_time <= $3", user.ID, startTime, endTime)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get data: %v", err), http.StatusInternalServerError)
	}

	response, err := json.Marshal(events)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal events data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(response)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		fmt.Printf("Error writing response: %v\n", err)
		return
	}
}
