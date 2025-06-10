package utils

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateJWT(UserID uuid.UUID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": UserID.String(),
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fiber.NewError(fiber.StatusInternalServerError, "JWT secret not set")
	}

	return token.SignedString([]byte(secret))
}
