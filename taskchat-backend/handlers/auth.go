package handlers

import (
	"regexp"
	"strings"
	"taskchat/database"
	model "taskchat/models"
	"taskchat/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest defines the registration input
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest defines the login input
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var emailRegex = regexp.MustCompile(`(?i)^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)

// Register creates a new user
func Register(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequest(c, "Invalid request body")
		}

		// Normalize input
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		req.Password = strings.TrimSpace(req.Password)

		if !emailRegex.MatchString(req.Email) {
			return utils.BadRequest(c, "Invalid email format")
		}
		if len(req.Password) < 6 {
			return utils.BadRequest(c, "Password must be at least 6 characters")
		}

		var existingUser model.User
		if err := db.Conn.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			return utils.Conflict(c, "Email already exists")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash password")
			return utils.InternalError(c, "Failed to process request")
		}

		user := model.User{
			ID:           uuid.New(),
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
		}

		if err := db.Conn.Create(&user).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create user")
			return utils.InternalError(c, "Failed to create user")
		}

		token, err := utils.GenerateJWT(user.ID, user.Email)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate JWT")
			return utils.InternalError(c, "Failed to generate token")
		}

		return utils.Created(c, fiber.Map{
			"token": token,
			"user":  fiber.Map{"id": user.ID, "email": user.Email},
		})
	}
}

// Login authenticates a user
func Login(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequest(c, "Invalid request body")
		}

		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		req.Password = strings.TrimSpace(req.Password)

		if req.Email == "" || req.Password == "" {
			return utils.BadRequest(c, "Email and password are required")
		}

		var user model.User
		if err := db.Conn.Where("email = ?", req.Email).First(&user).Error; err != nil {
			log.Error().Err(err).Msg("User not found")
			return utils.Unauthorized(c, "Invalid email or password")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			log.Error().Err(err).Msg("Password mismatch")
			return utils.Unauthorized(c, "Invalid email or password")
		}

		token, err := utils.GenerateJWT(user.ID, user.Email)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate JWT")
			return utils.InternalError(c, "Could not login")
		}

		return utils.Success(c, fiber.Map{
			"token": token,
			"user":  fiber.Map{"id": user.ID, "email": user.Email},
		})
	}
}

// ErrorHandler handles unexpected errors
func ErrorHandler(c *fiber.Ctx, err error) error {
	if e, ok := err.(*fiber.Error); ok {
		return c.Status(e.Code).JSON(fiber.Map{"success": false, "error": e.Message})
	}
	log.Error().Err(err).Msg("Unexpected error")
	return utils.InternalError(c, "Something went wrong")
}
