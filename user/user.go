package user

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ushiradineth/cron-be/auth"
)

func GetUserHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, err := GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), http.StatusInternalServerError)
		return
	}

	user.Password = ""

	response, err := json.Marshal(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal user data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func PostUserHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, _, err := DoesUserExist("", email, db)
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
		http.Error(w, fmt.Sprintf("Failed to hash password: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4)", uuid.New(), name, email, hashedPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert user data: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func PutUserHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, err := GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update user data: %v", err), http.StatusInternalServerError)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")

	_, count, err := DoesUserExist(user.ID.String(), email, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), http.StatusInternalServerError)
		return
	}

	if count > 1 {
		http.Error(w, fmt.Sprintf("Email already in use"), http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE users SET name=$1, email=$2 WHERE id=$3", name, email, user.ID.String())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update user data: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func PutUserPasswordHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, err := GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update user data: %v", err), http.StatusInternalServerError)
		return
	}

	password, err := auth.HashPassword(r.FormValue("password"))
	if err != nil {
		log.Fatalln(err)
	}

	_, err = db.Exec("UPDATE users SET password=$1 WHERE id=$2", password, user.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update user password: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	user, err := GetUserFromJWT(r, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete user data: %v", err), http.StatusInternalServerError)
		return
	}

	res, err := db.Exec("DELETE FROM users WHERE id=$1", user.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete user: %v", err), http.StatusBadRequest)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete user: %v", err), http.StatusBadRequest)
		return
	}

	if count == 0 {
		http.Error(w, fmt.Sprintf("User does not exist"), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type AuthenticateUserResponse struct {
	User
	AccessToken  string
	RefreshToken string
}

func AuthenticateUserHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := GetUser(email, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user data: %v", err), http.StatusInternalServerError)
		return
	}

	valid := auth.CheckPasswordHash(password, user.Password)

	if !valid {
		http.Error(w, fmt.Sprintf("Invalid Credentials"), http.StatusUnauthorized)
		return
	}

	accessTokenClaim := auth.UserClaim{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
	}

	accessToken, err := auth.NewAccessToken(accessTokenClaim)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate access token: %v", err), http.StatusInternalServerError)
		return
	}

	refreshTokenClaim := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
	}

	refreshToken, err := auth.NewRefreshToken(refreshTokenClaim)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate refresh token: %v", err), http.StatusInternalServerError)
		return
	}

	user.Password = ""

	response := AuthenticateUserResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal user data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(jsonResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

type RefreshTokenResponse struct {
	AccessToken string
}

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	accessToken, err := auth.GetJWT(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	refreshToken := r.FormValue("refresh_token")

	accessTokenClaim, err := auth.ParseAccessToken(accessToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing access token: %v", err), http.StatusUnauthorized)
		return
	}

	refreshTokenClaim, err := auth.ParseRefreshToken(refreshToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing refresh token: %v", err), http.StatusUnauthorized)
		return
	}

	if refreshTokenClaim.ExpiresAt < time.Now().Unix() {
		http.Error(w, fmt.Sprintf("Refresh Token has expired, Please Log in again"), http.StatusUnauthorized)
		return
	}

	if accessTokenClaim.StandardClaims.Valid() == nil {
		http.Error(w, fmt.Sprint("Access Token is valid"), http.StatusBadRequest)
		return
	}

	newAccessTokenClaim := auth.UserClaim{
		Id:    accessTokenClaim.Id,
		Name:  accessTokenClaim.Name,
		Email: accessTokenClaim.Email,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
	}

	newAccessToken, err := auth.NewAccessToken(newAccessTokenClaim)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Access Token: %v", err), http.StatusInternalServerError)
	}

	response := RefreshTokenResponse{
		AccessToken: newAccessToken,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal user data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(jsonResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
