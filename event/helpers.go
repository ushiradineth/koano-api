package event

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func GetEvent(id string, user_id string, db *sqlx.DB) (*Event, error) {
	event := Event{}

	err := db.Get(&event, "SELECT * FROM events WHERE id=$1 AND user_id=$2", id, user_id)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func DoesEventExist(id string, start_time string, end_time string, user_id string, db *sqlx.DB) (bool, error) {
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

	if id_uuid == uuid.Nil {
		query = "SELECT COUNT(*) FROM events WHERE id=$1"
		args = append(args, id_uuid)
	} else {
		query = "SELECT COUNT(*) FROM events WHERE start_time=$1 AND end_time=$2 AND user_id=$3"
		args = append(args, start_time, end_time, user_id_uuid)
	}

	db.Get(&event, query, args...)

	return event != 0, nil
}
