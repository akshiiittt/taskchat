package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateJWT(UserID uuid.UUID, email string) (string, error) {
	// setting up the data that we need to store in the jwt
	claims := jwt.MapClaims{
		"user_id": UserID.String(),
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
		"jti":     uuid.New().String(),
	}

	//create the token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT SECRET not set")
	}

	// signing the token with the secret
	return token.SignedString([]byte(secret))
}
