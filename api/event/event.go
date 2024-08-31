package event

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ushiradineth/cron-be/models"
	"github.com/ushiradineth/cron-be/util"
)

var validate = validator.New()

// @Summary		Get Event
// @Description	Get authenticated user's event based on the JWT and parameters sent with the request
// @Tags			Event
// @Produce		json
// @Param			Path	path		EventPathParams	true	"EventPathParams"
// @Success		200		{object}	util.Response{data=models.Event}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Security		BearerAuth
// @Router			/event/{event_id} [get]
func Get(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user := util.GetUserFromJWT(r, w, db)

	path := EventPathParams{
		EventID: r.PathValue("event_id"),
	}

	if err := validate.Struct(path); err != nil {
		util.HTTPError(w, http.StatusBadRequest, err.Error(), util.StatusFail)
		return
	}

	event, err := util.GetEvent(path.EventID, user.ID.String(), db)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	util.HTTPResponse(w, event)
}

// @Summary		Create Event
// @Description	Create Event based on the parameters sent with the request
// @Tags			Event
//
// @Accept			json
//
// @Produce		json
// @Param			Query	query		PostQueryParams	true	"PostQueryParams"
// @Success		200		{object}	util.Response{data=models.Event}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Security		BearerAuth
// @Router			/event [post]
func Post(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user := util.GetUserFromJWT(r, w, db)

	query := PostQueryParams{
		Title:    r.FormValue("title"),
		Timezone: r.FormValue("timezone"),
		Repeated: r.FormValue("repeated"),
		Start:    r.FormValue("start_time"),
		End:      r.FormValue("end_time"),
	}

	if err := validate.Struct(query); err != nil {
		util.HTTPError(w, http.StatusBadRequest, err.Error(), util.StatusFail)
		return
	}

	eventExists := util.DoesEventExist("", query.Start, query.End, user.ID.String(), db)
	if eventExists {
		util.HTTPError(w, http.StatusBadRequest, "Event already exists", util.StatusFail)
		return
	}

	parsedStart, err := time.Parse(time.RFC3339, query.Start)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	parsedEnd, err := time.Parse(time.RFC3339, query.End)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	if parsedStart.After(parsedEnd) {
		util.HTTPError(w, http.StatusBadRequest, "Start time must not be after end time", util.StatusFail)
		return
	}

	event := models.Event{
		ID:       uuid.New(),
		Title:    query.Title,
		Start:    parsedStart,
		End:      parsedEnd,
		UserID:   user.ID,
		Timezone: query.Timezone,
		Repeated: query.Repeated,
	}

	_, err = db.Exec("INSERT INTO events (id, title, start_time, end_time, user_id, timezone, repeated) VALUES ($1, $2, $3, $4, $5, $6, $7)", event.ID, event.Title, event.Start, event.End, event.UserID, event.Timezone, event.Repeated)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	util.HTTPResponse(w, event)
}

// @Summary		Update Event
// @Description	Update Event based on the parameters sent with the request
// @Tags			Event
//
// @Accept			json
// @Produce		json
// @Param			Path	path		EventPathParams	true	"EventPathParams"
// @Param			Query	query		PutQueryParams	true	"PutQueryParams"
// @Success		200		{object}	util.Response{data=models.Event}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Security		BearerAuth
// @Router			/event/{event_id} [put]
func Put(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user := util.GetUserFromJWT(r, w, db)

	path := EventPathParams{
		EventID: r.PathValue("event_id"),
	}

	query := PutQueryParams{
		Title:    r.FormValue("title"),
		Timezone: r.FormValue("timezone"),
		Repeated: r.FormValue("repeated"),
		Start:    r.FormValue("start_time"),
		End:      r.FormValue("end_time"),
	}

	if err := validate.Struct(path); err != nil {
		util.HTTPError(w, http.StatusBadRequest, err.Error(), util.StatusFail)
		return
	}

	if err := validate.Struct(query); err != nil {
		util.HTTPError(w, http.StatusBadRequest, err.Error(), util.StatusFail)
		return
	}

	existingEvent := util.DoesEventExist(path.EventID, query.Start, query.End, user.ID.String(), db)

	if !existingEvent {
		util.HTTPError(w, http.StatusBadRequest, "Event already exists", util.StatusFail)
		return
	}

	parsedStart, err := time.Parse(time.RFC3339, query.Start)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	parsedEnd, err := time.Parse(time.RFC3339, query.End)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	if parsedStart.After(parsedEnd) {
		util.HTTPError(w, http.StatusBadRequest, "Start time must not be after end time", util.StatusFail)
		return
	}

	parsedUUID, err := uuid.Parse(path.EventID)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	event := models.Event{
		ID:       parsedUUID,
		Title:    query.Title,
		Start:    parsedStart,
		End:      parsedEnd,
		UserID:   user.ID,
		Timezone: query.Timezone,
		Repeated: query.Repeated,
	}

	_, err = db.Exec("UPDATE events SET title=$1, start_time=$2, end_time=$3, timezone=$4, repeated=$5 WHERE id=$6 AND user_id=$7", event.Title, event.Start, event.End, event.Timezone, event.Repeated, event.ID, event.UserID.String())
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	util.HTTPResponse(w, event)
}

// @Summary		Delete Event
// @Description	Delete Event based on the parameters sent with the request
// @Tags			Event
// @Accept			json
// @Produce		json
// @Param			Path	path		EventPathParams	true	"EventPathParams"
// @Success		200		{object}	util.Response{data=string}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Security		BearerAuth
// @Router			/event/{event_id} [delete]
func Delete(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user := util.GetUserFromJWT(r, w, db)

	path := EventPathParams{
		EventID: r.PathValue("event_id"),
	}

	if err := validate.Struct(path); err != nil {
		util.HTTPError(w, http.StatusBadRequest, err.Error(), util.StatusFail)
		return
	}

	res, err := db.Exec("DELETE FROM events WHERE id=$1 AND user_id=$2", path.EventID, user.ID.String())
	if err != nil {
		util.HTTPError(w, http.StatusBadRequest, "Event does not exist", util.StatusFail)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		util.HTTPError(w, http.StatusBadRequest, "Event does not exist", util.StatusFail)
		return
	}

	if count == 0 {
		util.HTTPError(w, http.StatusBadRequest, "Event does not exist", util.StatusFail)
		return
	}

	util.HTTPResponse(w, "Event has been successfully deleted")
}

// @Summary		Get User Events
// @Description	Get authenticated user's event based on the JWT sent with the request
// @Tags			Event
// @Accept			json
// @Produce		json
// @Param			Query	query		GetUserEventsQueryParams	true	"GetUserEventsQueryParams"
// @Success		200		{object}	util.Response{data=[]models.Event}
// @Failure		400		{object}	util.Error
// @Failure		401		{object}	util.Error
// @Failure		500		{object}	util.Error
// @Security		BearerAuth
// @Router			/event/user [get]
func GetUserEvents(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user := util.GetUserFromJWT(r, w, db)

	query := GetUserEventsQueryParams{
		Start: r.FormValue("start_day"),
		End:   r.FormValue("end_day"),
	}

	if err := validate.Struct(query); err != nil {
		util.HTTPError(w, http.StatusBadRequest, err.Error(), util.StatusFail)
		return
	}

	parsedStart, err := time.Parse("2006-01-02", query.Start)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	parsedEnd, err := time.Parse("2006-01-02", query.End)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	events := []models.Event{}

	err = db.Select(&events, "SELECT * FROM events WHERE user_id=$1 AND start_time >= $2 AND start_time <= $3", user.ID, parsedStart, parsedEnd)
	if err != nil {
		util.HTTPError(w, http.StatusInternalServerError, err.Error(), util.StatusError)
		return
	}

	util.HTTPResponse(w, events)
}
