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

type API struct {
	db        *sqlx.DB
	validator *validator.Validate
}

func New(db *sqlx.DB, validator *validator.Validate) *API {
	return &API{
		db:        db,
		validator: validator,
	}
}

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
func (api *API) Get(w http.ResponseWriter, r *http.Request) {
	user := util.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	path := EventPathParams{
		EventID: r.PathValue("event_id"),
	}

	if err := api.validator.Struct(path); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	event, err := util.GetEvent(path.EventID, user.ID.String(), api.db)
	if err != nil {
		util.GenericServerError(w, err)
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
func (api *API) Post(w http.ResponseWriter, r *http.Request) {
	user := util.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	query := PostQueryParams{
		Title:     r.FormValue("title"),
		Timezone:  r.FormValue("timezone"),
		Repeated:  r.FormValue("repeated"),
		StartTime: r.FormValue("start_time"),
		EndTime:   r.FormValue("end_time"),
	}

	if err := api.validator.Struct(query); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	parsedStart, err := time.Parse(time.RFC3339, query.StartTime)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	parsedEnd, err := time.Parse(time.RFC3339, query.EndTime)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	eventExists := util.DoesEventExist("", parsedStart.String(), parsedEnd.String(), user.ID.String(), api.db)
	if eventExists {
		util.HTTPError(w, http.StatusBadRequest, "Event already exists", util.StatusFail)
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

	_, err = api.db.Exec("INSERT INTO events (id, title, start_time, end_time, user_id, timezone, repeated) VALUES ($1, $2, $3, $4, $5, $6, $7)", event.ID, event.Title, event.Start, event.End, event.UserID, event.Timezone, event.Repeated)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	util.HTTPResponse(w, event)
}

// @Summary		Update Event
// @Description	Update Event based on the parameters sent with the request
// @Tags			Event
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
func (api *API) Put(w http.ResponseWriter, r *http.Request) {
	user := util.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	path := EventPathParams{
		EventID: r.PathValue("event_id"),
	}

	query := PutQueryParams{
		Title:     r.FormValue("title"),
		Timezone:  r.FormValue("timezone"),
		Repeated:  r.FormValue("repeated"),
		StartTime: r.FormValue("start_time"),
		EndTime:   r.FormValue("end_time"),
	}

	if err := api.validator.Struct(path); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	if err := api.validator.Struct(query); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	existingEvent := util.DoesEventExist(path.EventID, query.StartTime, query.EndTime, user.ID.String(), api.db)

	if !existingEvent {
		util.HTTPError(w, http.StatusBadRequest, "Event already exists", util.StatusFail)
		return
	}

	parsedStart, err := time.Parse(time.RFC3339, query.StartTime)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	parsedEnd, err := time.Parse(time.RFC3339, query.EndTime)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	if parsedStart.After(parsedEnd) {
		util.HTTPError(w, http.StatusBadRequest, "Start time must not be after end time", util.StatusFail)
		return
	}

	parsedUUID, err := uuid.Parse(path.EventID)
	if err != nil {
		util.GenericServerError(w, err)
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

	_, err = api.db.Exec("UPDATE events SET title=$1, start_time=$2, end_time=$3, timezone=$4, repeated=$5 WHERE id=$6 AND user_id=$7", event.Title, event.Start, event.End, event.Timezone, event.Repeated, event.ID, event.UserID.String())
	if err != nil {
		util.GenericServerError(w, err)
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
func (api *API) Delete(w http.ResponseWriter, r *http.Request) {
	user := util.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	path := EventPathParams{
		EventID: r.PathValue("event_id"),
	}

	if err := api.validator.Struct(path); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	res, err := api.db.Exec("DELETE FROM events WHERE id=$1 AND user_id=$2", path.EventID, user.ID.String())
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
func (api *API) GetUserEvents(w http.ResponseWriter, r *http.Request) {
	user := util.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	query := GetUserEventsQueryParams{
		StartDay: r.FormValue("start_day"),
		EndDay:   r.FormValue("end_day"),
	}

	if err := api.validator.Struct(query); err != nil {
		util.GenericValidationError(w, err)
		return
	}

	parsedStart, err := time.Parse("2006-01-02", query.StartDay)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	parsedEnd, err := time.Parse("2006-01-02", query.EndDay)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	events := []models.Event{}

	err = api.db.Select(&events, "SELECT * FROM events WHERE user_id=$1 AND start_time >= $2 AND start_time <= $3", user.ID, parsedStart, parsedEnd)
	if err != nil {
		util.GenericServerError(w, err)
		return
	}

	util.HTTPResponse(w, events)
}
