package user

func GetUser(id_or_email string) (*User, error) {
	user := User{}

	err := DB.Get(&user, "SELECT * FROM user WHERE id=(?) OR email=(?)", id_or_email, id_or_email)
	if err != nil {
		return nil, err
	}

	user.Password = ""

	return &user, nil
}

func DoesUserExist(id string, email string) (bool, int, error) {
	user := 0

	err := DB.Get(&user, "SELECT COUNT(*) FROM user WHERE id=(?) OR email=(?)", id, email)
	if err != nil {
		return false, 0, err
	}

	return user != 0, user, nil
}
