package handlers

import (
	"strings"
	"taskchat/database"
	"taskchat/models"
	"taskchat/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog/log"
)

type NoteRequest struct {
	Title string `json:"title"`
}

var noteSanitizer = bluemonday.UGCPolicy()

// get userID function
func getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return uuid.UUID{}, utils.InternalError(c, "Invalid user id")
	}
	// return response
	return userID, nil
}

// get notes function
func GetNotes(c *fiber.Ctx) error {

	//get the userID from the JWT
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	// fetch all the notes for that particular user
	var notes []models.Note
	if err := database.DB.Where("user_id=?", userID).Find(&notes).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch notes")
		return utils.InternalError(c, "Failed to fetch notes")
	}

	// return response
	return utils.Success(c, fiber.Map{"notes": notes})
}

// create notes function
func CreateNote(c *fiber.Ctx) error {

	//get the userID from the JWT
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	//parse the request body
	var req NoteRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid req body ")
	}

	// clean and santize req
	req.Title = noteSanitizer.Sanitize(strings.TrimSpace(req.Title))
	if req.Title == "" || len(req.Title) > 100 {
		return utils.BadRequest(c, "Title is required and must under 100 characters")
	}

	// create note
	note := models.Note{
		ID:     uuid.New(),
		UserID: userID,
		Title:  req.Title,
	}

	// save note to database
	if err := database.DB.Create(&note).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create note")
		return utils.BadRequest(c, "Failed to create note")
	}

	// return response
	return utils.Created(c, note)
}

// update notes function
func UpdateNote(c *fiber.Ctx) error {

	//get the userID from the JWT
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	// get the note id from params
	noteID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "Invalid note id")
	}

	// parse req body
	var req NoteRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid req body ")
	}

	req.Title = noteSanitizer.Sanitize(strings.TrimSpace(req.Title))
	if req.Title == "" || len(req.Title) > 100 {
		return utils.BadRequest(c, "Title is required and mush under 100 characters")
	}

	// find the note from the db
	var note models.Note
	if err := database.DB.Where("id=? AND user_id=?", noteID, userID).First(&note).Error; err != nil {
		log.Error().Err(err).Msg("Note not found")
		return utils.NotFound(c, "Note not found")
	}

	//update the title
	note.Title = req.Title
	if err := database.DB.Save(&note).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create note")
		return utils.InternalError(c, "Failed to update note")
	}

	// return response
	return utils.Success(c, note)
}

// delete notes function
func DeleteNote(c *fiber.Ctx) error {

	//get the userID from the JWT
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	// get the note id from params
	noteID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "Invalid note id")
	}

	// find the note from the db
	var note models.Note
	if err := database.DB.Where("id=? AND user_id=?", noteID, userID).First(&note).Error; err != nil {
		log.Error().Err(err).Msg("Note not found")
		return utils.NotFound(c, "Note not found")
	}

	if err := database.DB.Where("note_id=?", noteID).Delete(&models.Task{}).Error; err != nil {
		log.Error().Err(err).Msg("Failed to delete tasks")
		return utils.InternalError(c, "Failed to delete tasks ")
	}

	if err := database.DB.Delete(&note).Error; err != nil {
		log.Error().Err(err).Msg("Failed to delete note")
		return utils.InternalError(c, "Failed to delete note")
	}

	// return response
	return utils.Success(c, fiber.Map{"message": "Note deleted successfully"})
}
