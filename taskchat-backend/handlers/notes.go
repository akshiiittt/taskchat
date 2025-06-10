package handlers

import (
	"log"
	"taskchat/database"
	"taskchat/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type NoteRequest struct {
	Title string `json:"title"`
}

func GetNotes(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid user id"})
	}

	var notes []models.Note
	if err := database.DB.Where("user_id=?", userID).Find(&notes).Error; err != nil {
		log.Println("Failed to fetch notes", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch notes"})
	}

	return c.JSON(fiber.Map{"notes": notes})
}

func CreateNote(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid user id"})
	}

	var req NoteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid req body "})
	}

	if req.Title == "" || len(req.Title) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title is required and must under 100 characters"})
	}

	note := models.Note{
		ID:     uuid.New(),
		UserID: userID,
		Title:  req.Title,
	}

	if err := database.DB.Create(&note).Error; err != nil {
		log.Println("Failed to create note", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create note"})
	}

	return c.JSON(note)
}

func UpdateNote(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid user id"})
	}

	noteID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid note id"})
	}

	var req NoteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid req body "})
	}

	if req.Title == "" || len(req.Title) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title is required and mush under 100 characters"})
	}

	var note models.Note
	if err := database.DB.Where("id=? AND user_id=?", noteID, userID).First(&note).Error; err != nil {
		log.Println("Failed to update note", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Note not found"})
	}

	note.Title = req.Title
	if err := database.DB.Save(&note).Error; err != nil {
		log.Println("Failed to create note", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update note"})
	}

	return c.JSON(note)
}

func DeleteNote(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid user id"})
	}

	noteID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid note id"})
	}

	var note models.Note
	if err := database.DB.Where("id=? AND user_id=?", noteID, userID).First(&note).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Note not found"})
	}

	if err := database.DB.Where("note_id=?", noteID).Delete(&models.Task{}).Error; err != nil {
		log.Println("Failed to delete tasks", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete tasks "})
	}

	if err := database.DB.Delete(&note).Error; err != nil {
		log.Println("Failed to delete note", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete note"})
	}

	return c.JSON(fiber.Map{"message": "Note deleted successfully"})
}
