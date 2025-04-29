package utils

import (
	"time"

	"github.com/golang-jwt/jwt"
)

var jwtKey = []byte("supersecretkey") // Replace for production!

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func GenerateJWT(username string) (string, time.Time, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, expirationTime, nil
}

func ValidateToken(tknStr string) (*Claims, error) {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	return claims, nil
}
