package handlers

import (
	"strings"
	"taskchat/database"
	model "taskchat/models"
	"taskchat/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog/log"
)

// TaskRequest defines task input
type TaskRequest struct {
	Title    string `json:"title,omitempty"`
	Status   string `json:"status,omitempty"`
	Priority string `json:"priority,omitempty"`
}

var taskSanitizer = bluemonday.UGCPolicy()

// GetPriorities fetches high-priority tasks
func GetPriorities(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return err
		}

		var tasks []model.Task
		if err := db.Conn.Where("user_id = ? AND priority = ?", userID, "high").Find(&tasks).Error; err != nil {
			log.Error().Err(err).Msg("Failed to fetch priority tasks")
			return utils.InternalError(c, "Failed to fetch priority tasks")
		}

		return utils.Success(c, fiber.Map{"tasks": tasks})
	}
}

// GetTasks fetches tasks for a note
func GetTasks(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return err
		}

		noteID, err := uuid.Parse(c.Params("note_id"))
		if err != nil {
			return utils.BadRequest(c, "Invalid note ID")
		}

		var note model.Note
		if err := db.Conn.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
			log.Error().Err(err).Msg("Note not found")
			return utils.NotFound(c, "Note not found")
		}

		var tasks []model.Task
		if err := db.Conn.Where("note_id = ? AND user_id = ?", noteID, userID).Find(&tasks).Error; err != nil {
			log.Error().Err(err).Msg("Failed to fetch tasks")
			return utils.InternalError(c, "Failed to fetch tasks")
		}

		return utils.Success(c, fiber.Map{"tasks": tasks})
	}
}

// CreateTask creates a new task
func CreateTask(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return err
		}

		noteID, err := uuid.Parse(c.Params("note_id"))
		if err != nil {
			return utils.BadRequest(c, "Invalid note ID")
		}

		var note model.Note
		if err := db.Conn.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
			log.Error().Err(err).Msg("Note not found")
			return utils.NotFound(c, "Note not found")
		}

		var req TaskRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequest(c, "Invalid request body")
		}

		req.Title = taskSanitizer.Sanitize(strings.TrimSpace(req.Title))
		if req.Title == "" || len(req.Title) > 255 {
			return utils.BadRequest(c, "Title is required and must be under 255 characters")
		}

		if req.Priority != "" && req.Priority != "high" {
			return utils.BadRequest(c, "Priority must be 'high' or empty")
		}

		task := model.Task{
			ID:       uuid.New(),
			NoteID:   noteID,
			UserID:   userID,
			Title:    req.Title,
			Status:   string(model.StatusPending),
			Priority: req.Priority,
		}

		if err := db.Conn.Create(&task).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create task")
			return utils.InternalError(c, "Failed to create task")
		}

		return utils.Created(c, task)
	}
}

// UpdateTask updates a task
func UpdateTask(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return err
		}

		taskID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return utils.BadRequest(c, "Invalid task ID")
		}

		var req TaskRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequest(c, "Invalid request body")
		}

		var task model.Task
		if err := db.Conn.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
			log.Error().Err(err).Msg("Task not found")
			return utils.NotFound(c, "Task not found")
		}

		if req.Title != "" {
			req.Title = taskSanitizer.Sanitize(strings.TrimSpace(req.Title))
			if len(req.Title) > 255 {
				return utils.BadRequest(c, "Title must be under 255 characters")
			}
			task.Title = req.Title
		}

		if req.Status != "" {
			if req.Status != string(model.StatusPending) && req.Status != string(model.StatusCompleted) {
				return utils.BadRequest(c, "Status must be 'pending' or 'completed'")
			}
			task.Status = req.Status
		}

		if req.Priority != "" && req.Priority != "high" {
			return utils.BadRequest(c, "Priority must be 'high' or empty")
		}
		task.Priority = req.Priority

		if err := db.Conn.Save(&task).Error; err != nil {
			log.Error().Err(err).Msg("Failed to update task")
			return utils.InternalError(c, "Failed to update task")
		}

		return utils.Success(c, task)
	}
}

// DeleteTask deletes a task
func DeleteTask(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return err
		}

		taskID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return utils.BadRequest(c, "Invalid task ID")
		}

		var task model.Task
		if err := db.Conn.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
			log.Error().Err(err).Msg("Task not found")
			return utils.NotFound(c, "Task not found")
		}

		if err := db.Conn.Delete(&task).Error; err != nil {
			log.Error().Err(err).Msg("Failed to delete task")
			return utils.InternalError(c, "Failed to delete task")
		}

		return utils.Success(c, fiber.Map{"message": "Task deleted successfully"})
	}
}
