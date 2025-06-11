package middleware

import (
	"log"
	"os"
	"strings"
	"taskchat/utils"

	"github.com/gofiber/fiber/v2"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func AuthMiddleware(c *fiber.Ctx) error {
	// get authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return utils.Unauthorized(c, "Authorization header missing")
	}

	// check if the header is the bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return utils.Unauthorized(c, "Authorization header must be bearer token")
	}

	// parse the token
	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")

		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		log.Println("Failed to parse JWT", err)
		return utils.Unauthorized(c, "Invalid token")
	}

	// get the user id from the token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIdStr, ok := claims["user_id"].(string)
		if !ok {
			return utils.Unauthorized(c, "Invalid user_id in token")
		}

		userID, err := uuid.Parse(userIdStr)
		if err != nil {
			return utils.Unauthorized(c, "Invalid user_id in token")
		}

		// stroing the user id in the req body
		c.Locals("user_id", userID)
		return c.Next()
	}

	return utils.Unauthorized(c, "Invalid token")

}
