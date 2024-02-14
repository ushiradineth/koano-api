package user

import (
	"auth"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("user_id")
	fmt.Fprintf(w, "GET user with id=%v\n", id)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "POST user\n")

	name := r.FormValue("name")
	email := r.FormValue("email")
	password, err := auth.HashPassword(r.FormValue("password"))
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprintf(w, "Name = %s\n", name)
	fmt.Fprintf(w, "Email = %s\n", email)
	fmt.Fprintf(w, "Password = %s\n", password)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("user_id")
	fmt.Fprintf(w, "PUT user with id=%v\n", id)

	name := r.FormValue("name")
	email := r.FormValue("email")
	fmt.Fprintf(w, "Name = %s\n", name)
	fmt.Fprintf(w, "Email = %s\n", email)
}

func UpdateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("user_id")
	fmt.Fprintf(w, "PUT user with id=%v\n", id)

	password := r.FormValue("password")
	fmt.Fprintf(w, "Password = %s\n", password)
}

// func AuthenticateUser(w http.ResponseWriter, r *http.Request) {
// 	email := r.FormValue("email")
// 	password := r.FormValue("password")

// 	fmt.Fprintf(w, "Password = %s\n", password)
// }
