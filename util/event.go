package util

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/ushiradineth/cron-be/models"
)

func GetEvent(id string, user_id string, db *sqlx.DB) (*models.Event, error) {
	event := models.Event{}

	err := db.Get(&event, "SELECT * FROM events WHERE id=$1 AND user_id=$2", id, user_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Event not found")
		}
		return nil, err
	}

	return &event, nil
}

func DoesEventExist(id string, start_time string, end_time string, user_id string, db *sqlx.DB) bool {
	event := 0
	var query string
	var args []interface{}

	id_uuid, err := uuid.Parse(id)
	if err != nil {
		id_uuid = uuid.Nil
	}

	user_id_uuid, err := uuid.Parse(user_id)
	if err != nil {
		user_id_uuid = uuid.Nil
	}

	if id_uuid != uuid.Nil {
		query = "SELECT COUNT(*) FROM events WHERE id=$1 AND user_id=$2"
		args = append(args, id_uuid, user_id_uuid)
	} else {
		query = "SELECT COUNT(*) FROM events WHERE start_time=$1 AND end_time=$2 AND user_id=$3"
		args = append(args, start_time, end_time, user_id_uuid)
	}

	db.Get(&event, query, args...)

	return event != 0
}
