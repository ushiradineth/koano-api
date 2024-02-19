package auth

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type UserClaim struct {
	Id    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	jwt.StandardClaims
}

func NewAccessToken(claims UserClaim) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return accessToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func NewRefreshToken(claims jwt.StandardClaims) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return refreshToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func ParseAccessToken(accessToken string) (*UserClaim, error) {
	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, &UserClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	if parsedAccessToken == nil {
		return nil, errors.New("unable to parse access token")
	}

	claims, ok := parsedAccessToken.Claims.(*UserClaim)
	if !ok || !parsedAccessToken.Valid {
		return nil, errors.New("invalid access token")
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("access token has expired")
	}

	return claims, nil
}

func ParseRefreshToken(refreshToken string) (*jwt.StandardClaims, error) {
	parsedRefreshToken, err := jwt.ParseWithClaims(refreshToken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	if parsedRefreshToken == nil {
		return nil, errors.New("unable to parse refresh token")
	}

	claims, ok := parsedRefreshToken.Claims.(*jwt.StandardClaims)
	if !ok || !parsedRefreshToken.Valid {
		return nil, errors.New("invalid refresh token")
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("refresh token has expired")
	}

	return claims, nil
}

func GetJWT(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization header is missing")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("Invalid Authorization header format")
	}

	return parts[1], nil
}
