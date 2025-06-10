package handlers

import (
	"log"
	"taskchat/database"
	"taskchat/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TaskRequest struct {
	Title    string `json:"title,omitempty"`
	Status   string `json:"status,omitempty"`
	Priority string `json:"priority,omitempty"`
}

func GetPriorities(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid user id"})
	}

	var tasks []models.Task
	if err := database.DB.Where("user_id=? AND priority !=?", userID, "none").Find(&tasks).Error; err != nil {
		log.Println("Failed to fetch priority tasks", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch priorities tasks"})
	}

	return c.JSON(fiber.Map{"tasks": tasks})

}

func GetTasks(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid user id"})
	}

	noteID, err := uuid.Parse(c.Params("note_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid note id"})
	}

	var note models.Note
	if err := database.DB.Where("id=? AND user_id=?", noteID, userID).First(&note).Error; err != nil {
		log.Println("Note not found", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Note not found"})
	}

	var tasks []models.Task
	if err := database.DB.Where("note_id=? AND user_id=?", noteID, userID).Find(&tasks).Error; err != nil {
		log.Println("Failed to fetch notes", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch tasks"})
	}

	return c.JSON(fiber.Map{"tasks": tasks})
}

func CreateTask(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid user id"})
	}

	noteID, err := uuid.Parse(c.Params("note_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid note id"})
	}

	var note models.Note
	if err := database.DB.Where("id=? AND user_id=?", noteID, userID).First(&note).Error; err != nil {
		log.Println("Note not found", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Note not found"})
	}

	var req TaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid req body "})
	}

	if req.Title == "" || len(req.Title) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title is required and must under 500 characters"})
	}

	task := models.Task{
		ID:       uuid.New(),
		NoteID:   noteID,
		UserID:   userID,
		Title:    req.Title,
		Status:   "pending",
		Priority: "none",
	}

	if err := database.DB.Create(&task).Error; err != nil {
		log.Println("Failed to create task", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create task"})
	}

	return c.JSON(task)
}

func UpdateTask(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid user id"})
	}

	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task id"})
	}

	var req TaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid req body "})
	}

	var task models.Task
	if err := database.DB.Where("id=? AND user_id=?", taskID, userID).First(&task).Error; err != nil {
		log.Println("Failed to update task", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Task not found"})
	}

	if req.Status != "" {
		if req.Status != "pending" && req.Status != "completed" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status"})
		}
		task.Status = req.Status
	}

	if req.Priority != "" {
		if req.Priority != "none" && req.Priority != "low" && req.Priority != "medium" && req.Priority != "high" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid priority"})
		}
		task.Priority = req.Priority
	}

	if err := database.DB.Save(&task).Error; err != nil {
		log.Println("Failed to update task", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update task"})
	}

	return c.JSON(task)
}

func DeleteTask(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid user id"})
	}

	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task id"})
	}

	var task models.Task
	if err := database.DB.Where("id=? AND user_id=?", taskID, userID).First(&task).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Task not found"})
	}

	if err := database.DB.Delete(&task).Error; err != nil {
		log.Println("Failed to delete task", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete task"})
	}

	return c.JSON(fiber.Map{"message": "Task deleted successfully"})
}
