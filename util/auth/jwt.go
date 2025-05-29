package auth

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/ushiradineth/koano-api/util/response"
)

type UserClaim struct {
	Id    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	jwt.StandardClaims
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

func NewAccessToken(id uuid.UUID, name string, email string) (string, int64, int64, error) {
	expiresIn := int64(15 * 60)
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second).Unix()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaim{
		Id:    id,
		Name:  name,
		Email: email,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: expiresAt,
		},
	})

	signedToken, err := accessToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", 0, 0, err
	}

	return signedToken, expiresIn, expiresAt, nil
}

func NewRefreshToken() (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
	})

	return refreshToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func ParseAccessToken(w http.ResponseWriter, accessToken string) *UserClaim {
	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, &UserClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		response.GenericUnauthenticatedError(w)
		return nil
	}

	claims, ok := parsedAccessToken.Claims.(*UserClaim)
	if ok && parsedAccessToken.Valid {
		return claims
	}

	return nil
}

func ParseRefreshToken(w http.ResponseWriter, refreshToken string) *jwt.StandardClaims {
	parsedRefreshToken, err := jwt.ParseWithClaims(refreshToken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		response.GenericUnauthenticatedError(w)
		return nil
	}

	claims, ok := parsedRefreshToken.Claims.(*jwt.StandardClaims)
	if ok && parsedRefreshToken.Valid {
		return claims
	}

	return nil
}

func ParseExpiredAccessToken(w http.ResponseWriter, accessToken string) *UserClaim {
	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, &UserClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err == nil || !strings.Contains(err.Error(), "token is expired by") {
		response.GenericBadRequestError(w, errors.New("Token is valid"))
		return nil
	}

	claims, ok := parsedAccessToken.Claims.(*UserClaim)
	if ok && parsedAccessToken.Valid {
		response.GenericBadRequestError(w, errors.New("Token is valid"))
		return nil
	}

	return claims
}
