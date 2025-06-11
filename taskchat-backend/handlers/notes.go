package handlers

import (
	"fmt"
	"strings"
	"taskchat/database"
	model "taskchat/models"
	"taskchat/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// NoteRequest defines note input
type NoteRequest struct {
	Title string `json:"title"`
}

var noteSanitizer = bluemonday.UGCPolicy()

// getUserID extracts user ID from context
func getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return uuid.UUID{}, utils.InternalError(c, "Invalid user ID")
	}
	return userID, nil
}

// GetNotes fetches all user notes
func GetNotes(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return err
		}

		var notes []model.Note
		if err := db.Conn.Where("user_id = ?", userID).Find(&notes).Error; err != nil {
			log.Error().Err(err).Msg("Failed to fetch notes")
			return utils.InternalError(c, "Failed to fetch notes")
		}

		return utils.Success(c, fiber.Map{"notes": notes})
	}
}

// CreateNote creates a new note
func CreateNote(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return err
		}

		var req NoteRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequest(c, "Invalid request body")
		}

		req.Title = noteSanitizer.Sanitize(strings.TrimSpace(req.Title))
		if req.Title == "" || len(req.Title) > 100 {
			return utils.BadRequest(c, "Title is required and must be under 100 characters")
		}

		note := model.Note{
			ID:     uuid.New(),
			UserID: userID,
			Title:  req.Title,
		}

		if err := db.Conn.Create(&note).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create note")
			return utils.InternalError(c, "Failed to create note")
		}

		return utils.Created(c, note)
	}
}

// UpdateNote updates a note
func UpdateNote(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return err
		}

		noteID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return utils.BadRequest(c, "Invalid note ID")
		}

		var req NoteRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequest(c, "Invalid request body")
		}

		req.Title = noteSanitizer.Sanitize(strings.TrimSpace(req.Title))
		if req.Title == "" || len(req.Title) > 100 {
			return utils.BadRequest(c, "Title is required and must be under 100 characters")
		}

		var note model.Note
		if err := db.Conn.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
			log.Error().Err(err).Msg("Note not found")
			return utils.NotFound(c, "Note not found")
		}

		note.Title = req.Title
		if err := db.Conn.Save(&note).Error; err != nil {
			log.Error().Err(err).Msg("Failed to update note")
			return utils.InternalError(c, "Failed to update note")
		}

		return utils.Success(c, note)
	}
}

// DeleteNote deletes a note and its tasks
func DeleteNote(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return err
		}

		noteID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return utils.BadRequest(c, "Invalid note ID")
		}

		var note model.Note
		if err := db.Conn.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
			log.Error().Err(err).Msg("Note not found")
			return utils.NotFound(c, "Note not found")
		}

		err = db.Conn.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("note_id = ?", noteID).Delete(&model.Task{}).Error; err != nil {
				log.Error().Err(err).Msg("Failed to delete tasks")
				return fmt.Errorf("failed to delete tasks: %v", err)
			}
			if err := tx.Delete(&note).Error; err != nil {
				log.Error().Err(err).Msg("Failed to delete note")
				return fmt.Errorf("failed to delete note: %v", err)
			}
			return nil
		})
		if err != nil {
			return utils.InternalError(c, "Failed to delete note")
		}

		return utils.Success(c, fiber.Map{"message": "Note deleted successfully"})
	}
}
