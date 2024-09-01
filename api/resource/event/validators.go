package event

type UserPathParams struct {
	UserID string `form:"user_id" validate:"required,uuid"`
}

type EventPathParams struct {
	EventID string `form:"event_id" validate:"required,uuid"`
}

type PostQueryParams struct {
	Title     string `form:"title" validate:"required"`
	Timezone  string `form:"timezone" validate:"required,timezone"`
	Repeated  string `form:"repeated" validate:"required"`
	StartTime string `form:"start_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	EndTime   string `form:"end_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

type PutQueryParams struct {
	Title     string `form:"title" validate:"required"`
	Timezone  string `form:"timezone" validate:"required,timezone"`
	Repeated  string `form:"repeated" validate:"required"`
	StartTime string `form:"start_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	EndTime   string `form:"end_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

type GetUserEventsQueryParams struct {
	StartDay string `form:"start_day" validate:"required,datetime=2006-01-02"`
	EndDay   string `form:"end_day" validate:"required,datetime=2006-01-02"`
}
