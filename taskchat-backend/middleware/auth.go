package middleware

import (
	"os"
	"strings"
	"taskchat/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware verifies JWT tokens
func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return utils.Unauthorized(c, "Authorization header missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return utils.Unauthorized(c, "Authorization header must be Bearer token")
	}

	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse JWT")
		return utils.Unauthorized(c, "Invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return utils.Unauthorized(c, "Invalid user ID in token")
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			log.Error().Err(err).Msg("Invalid user ID format")
			return utils.Unauthorized(c, "Invalid user ID in token")
		}
		c.Locals("user_id", userID)
		return c.Next()
	}

	return utils.Unauthorized(c, "Invalid token")
}
