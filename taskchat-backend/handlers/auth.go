package handlers

import (
	"regexp"
	"strings"
	"taskchat/database"
	"taskchat/models"
	"taskchat/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// type of schema for registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// type of schema for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// regex for the email
var emailRegex = regexp.MustCompile(`(?i)^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)

// function for the signup
func Register(c *fiber.Ctx) error {

	// making the varible for the like storing the req body
	var req RegisterRequest

	// parsing the req into the req variable
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	// normalising the req
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	//checking the email is matches the regex or not
	if !emailRegex.MatchString(req.Email) {
		return utils.BadRequest(c, "Invalid email format")
	}

	if len(req.Password) < 6 {
		return utils.BadRequest(c, "Password must be atleast 6 characters")
	}

	// checking if the email already exits or not creating a variable
	var existingEmail models.User
	// checks in db is the email is present than it attaches to the varibles, .error tells the there is no error while running the query
	if err := database.DB.Where("email=?", req.Email).First(&existingEmail).Error; err == nil {
		return utils.Conflict(c, "Email alreasy exits")
	}

	// byte converts the data into single words and than hashes it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return utils.InternalError(c, "Failed to process request")
	}

	// creating a user model
	user := models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	// creating and saving the created user
	err = database.DB.Create(&user).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to create users")
		return utils.InternalError(c, "Failed to create user ")
	}

	//gemerating token
	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT")
		return utils.Conflict(c, "Failed to generate token")
	}

	// return response
	return utils.Success(c, fiber.Map{"token": token, "user": fiber.Map{"id": user.ID, "email": user.Email}})

}

func Login(c *fiber.Ctx) error {

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		return utils.BadRequest(c, "Email and password are required")
	}

	var user models.User
	if err := database.DB.Where("email=?", req.Email).First(&user).Error; err != nil {
		log.Error().Err(err).Msg("User not found in DB")
		return utils.Unauthorized(c, "Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Error().Err(err).Msg("Password missmatch")
		return utils.Unauthorized(c, "Invalid email or password")
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT")
		return utils.InternalError(c, "Could not login, try again")
	}

	// return response
	return utils.Created(c, fiber.Map{"token": token, "user": fiber.Map{"id": user.ID, "email": user.Email}})

}

func ErrorHandler(c *fiber.Ctx, err error) error {
	if e, ok := err.(*fiber.Error); ok {
		return c.Status(e.Code).JSON(fiber.Map{"success": false, "error": e.Message})
	}

	log.Error().Err(err).Msg("unexpected error")

	// return response
	return utils.InternalError(c, "Error by error handler")
}
