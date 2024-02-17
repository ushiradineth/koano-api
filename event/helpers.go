package event

func GetEvent(id string) (*Event, error) {
	event := Event{}

	err := DB.Get(&event, "SELECT * FROM event WHERE id=(?)", id)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func DoesEventExist(id string, start string, end string, user_id string) (bool, error) {
	event := 0

	err := DB.Get(&event, "SELECT COUNT(*) FROM event WHERE id=(?) OR (start=(?) AND end=(?) AND user_id=(?))", id, start, end, user_id)
	if err != nil {
		return false, err
	}

	return event != 0, nil
}
