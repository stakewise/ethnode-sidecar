package clients

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type AuthorizationMethod string

const (
	None AuthorizationMethod = "bearer"
	// Bearer Basic  AuthorizationMethod = "basic"
	Bearer AuthorizationMethod = "bearer"
)

func CreateJWTAuthToken(jwtSecret string) (string, error) {
	secret, err := hex.DecodeString(strings.TrimSpace(jwtSecret))
	if err != nil {
		return "", err
	}

	// TODO: caching strategy
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	})
	tokenString, err := token.SignedString(secret)
	return tokenString, err
}
