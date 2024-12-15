package event

type UserPathParams struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

type EventPathParams struct {
	EventID string `json:"event_id" validate:"required,uuid"`
}

type EventBodyParams struct {
	Title     string `json:"title" validate:"required"`
	Timezone  string `json:"timezone" validate:"required,timezone"`
	Repeated  string `json:"repeated" validate:"required,oneof=never daily weekly monthly yearly"`
	StartTime string `json:"start_time" validate:"required,datetime=2006-01-02T15:04:05Z"`
	EndTime   string `json:"end_time" validate:"required,datetime=2006-01-02T15:04:05Z"`
}

type GetUserEventsBodyParams struct {
	StartDay string `json:"start_day" validate:"required,datetime=2006-01-02"`
	EndDay   string `json:"end_day" validate:"required,datetime=2006-01-02"`
}
