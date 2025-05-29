package event

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ushiradineth/koano-api/models"
	"github.com/ushiradineth/koano-api/util/event"
	logger "github.com/ushiradineth/koano-api/util/log"
	"github.com/ushiradineth/koano-api/util/response"
	"github.com/ushiradineth/koano-api/util/user"
)

type API struct {
	db        *sqlx.DB
	validator *validator.Validate
	log       *logger.Logger
}

func New(db *sqlx.DB, validator *validator.Validate, log *logger.Logger) *API {
	return &API{
		db:        db,
		validator: validator,
		log:       log,
	}
}

// @Summary		Get Event by ID
// @Description	Get authenticated user's event based on the JWT and event ID sent with the request
// @Tags			Event
// @Produce		json
// @Param			Path	path		EventPathParams	true	"EventPathParams"
// @Success		200		{object}	response.Response{data=models.Event}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Security		BearerAuth
// @Router			/events/{event_id} [get]
func (api *API) Get(w http.ResponseWriter, r *http.Request) {
	path := EventPathParams{
		EventID: r.PathValue("event_id"),
	}

	if err := api.validator.Struct(path); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	user := user.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	event := event.GetEvent(w, path.EventID, user.ID.String(), api.db)
	if event == nil {
		return
	}

	api.log.Info.Printf("Event %s has been retrieved by user %s", path.EventID, user.ID)

	response.HTTPResponse(w, event)
}

// @Summary		Create Event
// @Description	Create Event based on the parameters sent with the request
// @Tags			Event
// @Accept			json
// @Produce		json
// @Param			Body	body		EventBodyParams	true	"EventBodyParams"
// @Success		200		{object}	response.Response{data=models.Event}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Security		BearerAuth
// @Router			/events [post]
func (api *API) Post(w http.ResponseWriter, r *http.Request) {
	var body EventBodyParams
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	if err := api.validator.Struct(body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	user := user.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	eventExists := event.DoesEventExist("", body.StartTime, body.EndTime, user.ID.String(), api.db)

	parsedStart, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	parsedEnd, err := time.Parse(time.RFC3339, body.EndTime)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	if eventExists {
		response.HTTPError(w, http.StatusBadRequest, "Event already exists", response.StatusFail)
		return
	}

	if parsedStart.After(parsedEnd) {
		response.HTTPError(w, http.StatusBadRequest, "Start time must not be after end time", response.StatusFail)
		return
	}

	eventData := models.Event{
		ID:       uuid.New(),
		Title:    body.Title,
		Start:    parsedStart,
		End:      parsedEnd,
		UserID:   user.ID,
		Timezone: body.Timezone,
		Repeated: body.Repeated,
	}

	var event models.Event
	err = api.db.Get(&event, "INSERT INTO events (id, title, start_time, end_time, user_id, timezone, repeated) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *", eventData.ID, eventData.Title, eventData.Start, eventData.End, eventData.UserID, eventData.Timezone, eventData.Repeated)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	api.log.Info.Printf("Event %s has been created by user %s", event.ID, event.UserID)

	response.HTTPResponse(w, event)
}

// @Summary		Update Event
// @Description	Update Event based on the parameters sent with the request
// @Tags			Event
// @Accept			json
// @Produce		json
// @Param			Path	path		EventPathParams	true	"EventPathParams"
// @Param			Body	body		EventBodyParams	true	"EventBodyParams"
// @Success		200		{object}	response.Response{data=models.Event}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Security		BearerAuth
// @Router			/events/{event_id} [put]
func (api *API) Put(w http.ResponseWriter, r *http.Request) {
	path := EventPathParams{
		EventID: r.PathValue("event_id"),
	}

	if err := api.validator.Struct(path); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	var body EventBodyParams
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	if err := api.validator.Struct(body); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	user := user.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	existingEvent := event.GetEvent(w, path.EventID, user.ID.String(), api.db)
	if existingEvent == nil {
		response.HTTPError(w, http.StatusBadRequest, "Event does not exists", response.StatusFail)
		return
	}

	parsedStart, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	parsedEnd, err := time.Parse(time.RFC3339, body.EndTime)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	if parsedStart.After(parsedEnd) {
		response.HTTPError(w, http.StatusBadRequest, "Start time must not be after end time", response.StatusFail)
		return
	}

	parsedUUID, err := uuid.Parse(path.EventID)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	eventData := models.Event{
		ID:       parsedUUID,
		Title:    body.Title,
		Start:    parsedStart,
		End:      parsedEnd,
		UserID:   user.ID,
		Timezone: body.Timezone,
		Repeated: body.Repeated,
	}

	var event models.Event
	err = api.db.Get(&event, "UPDATE events SET title=$1, start_time=$2, end_time=$3, timezone=$4, repeated=$5 WHERE id=$6 AND user_id=$7 RETURNING *", eventData.Title, eventData.Start, eventData.End, eventData.Timezone, eventData.Repeated, eventData.ID, eventData.UserID.String())
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	api.log.Info.Printf("Event %s has been updated by user %s", event.ID, event.UserID)

	response.HTTPResponse(w, event)
}

// @Summary		Delete Event
// @Description	Delete Event based on the parameters sent with the request
// @Tags			Event
// @Produce		json
// @Param			Path	path		EventPathParams	true	"EventPathParams"
// @Success		200		{object}	response.Response{data=string}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Security		BearerAuth
// @Router			/events/{event_id} [delete]
func (api *API) Delete(w http.ResponseWriter, r *http.Request) {
	path := EventPathParams{
		EventID: r.PathValue("event_id"),
	}

	if err := api.validator.Struct(path); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	user := user.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	res, err := api.db.Exec("UPDATE events SET active=false, deleted_at=$1 WHERE id=$2 AND user_id=$3", time.Now(), path.EventID, user.ID.String())
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	if count == 0 {
		response.GenericBadRequestError(w, fmt.Errorf("Event does not exist"))
		return
	}

	api.log.Info.Printf("Event %s has been deleted by user %s", path.EventID, user.ID)

	response.HTTPResponse(w, "Event has been successfully deleted")
}

// @Summary		Get User Events
// @Description	Get authenticated user's event based on the JWT sent with the request
// @Tags			Event
// @Accept			x-www-form-urlencoded
// @Produce		json
// @Param			Query	query		GetUserEventsQueryParams	true	"GetUserEventsQueryParams"
// @Success		200		{object}	response.Response{data=[]models.Event}
// @Failure		400		{object}	response.Error
// @Failure		401		{object}	response.Error
// @Failure		500		{object}	response.Error
// @Security		BearerAuth
// @Router			/events [get]
func (api *API) GetUserEvents(w http.ResponseWriter, r *http.Request) {
	query := GetUserEventsQueryParams{
		StartDay: r.FormValue("start_day"),
		EndDay:   r.FormValue("end_day"),
	}

	if err := api.validator.Struct(query); err != nil {
		response.GenericValidationError(w, err)
		return
	}

	user := user.GetUserFromJWT(r, w, api.db)
	if user == nil {
		return
	}

	parsedStart, err := time.Parse("2006-01-02", query.StartDay)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	parsedEnd, err := time.Parse("2006-01-02", query.EndDay)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	events := []models.Event{}

	err = api.db.Select(&events, "SELECT * FROM events WHERE user_id=$1 AND start_time >= $2 AND start_time <= $3 AND active=true", user.ID, parsedStart, parsedEnd)
	if err != nil {
		response.GenericServerError(w, err)
		return
	}

	api.log.Info.Printf("Events for user %s have been retrieved", user.ID)

	response.HTTPResponse(w, events)
}
