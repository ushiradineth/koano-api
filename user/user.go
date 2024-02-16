package user

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ushiradineth/cron-be/auth"
)

var DB *sqlx.DB

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("user_id")

	user, err := GetUser(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal user data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	w.WriteHeader(http.StatusOK)
}

func PostUserHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, _, err := DoesUserExist("", email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), http.StatusInternalServerError)
		return
	}

	if user {
		http.Error(w, fmt.Sprintf("User already exists"), http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = DB.Exec("INSERT INTO user (id, name, email, password) VALUES (?, ?, ?, ?)", uuid.New(), name, email, hashedPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert user data: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func PutUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("user_id")
	name := r.FormValue("name")
	email := r.FormValue("email")

	user, count, err := DoesUserExist(id, email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), http.StatusInternalServerError)
		return
	}

	if !user {
		http.Error(w, fmt.Sprintf("User doesn't exist"), http.StatusBadRequest)
		return
	}

	if count > 1 {
		http.Error(w, fmt.Sprintf("Email already in use"), http.StatusBadRequest)
		return
	}

	password, err := auth.HashPassword(r.FormValue("password"))
	if err != nil {
		log.Fatalln(err)
	}

	_, err = DB.Exec("UPDATE user SET name=(?), email=(?), password=(?) WHERE id=(?)", name, email, password, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert user data: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("user_id")

	res, err := DB.Exec("DELETE FROM user WHERE id=(?)", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("User doesn't exist"), http.StatusBadRequest)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("User doesn't exist"), http.StatusBadRequest)
		return
	}

	if count == 0 {
		http.Error(w, fmt.Sprintf("User doesn't exist"), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func AuthenticateUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := GetUser(email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), http.StatusInternalServerError)
		return
	}

	valid := auth.CheckPasswordHash(password, user.Password)

	if !valid {
		http.Error(w, fmt.Sprintf("Unauthorized"), http.StatusUnauthorized)
		return
	}

	response, err := json.Marshal(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal user data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	w.WriteHeader(http.StatusOK)
}
