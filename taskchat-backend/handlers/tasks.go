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

type TaskRequest struct {
	Title    string `json:"title,omitempty"`
	Status   string `json:"status,omitempty"`
	Priority string `json:"priority,omitempty"`
}

var taskSanitizer = bluemonday.UGCPolicy()

// get priorites tasks
func GetPriorities(c *fiber.Ctx) error {

	//get the userID from the JWT
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	// get all the priority tasks
	var tasks []models.Task
	if err := database.DB.Where("user_id=? AND priority =?", userID, "high").Find(&tasks).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch priority tasks")
		return utils.InternalError(c, "Failed to fetch priorities tasks")
	}

	// return response
	return utils.Success(c, fiber.Map{"tasks": tasks})

}

// get all tasks function
func GetTasks(c *fiber.Ctx) error {

	//get the userID from the JWT
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	// get note id from params
	noteID, err := uuid.Parse(c.Params("note_id"))
	if err != nil {
		return utils.BadRequest(c, "Invalid note id")
	}

	// check if note exists
	var note models.Note
	if err := database.DB.Where("id=? AND user_id=?", noteID, userID).First(&note).Error; err != nil {
		log.Error().Err(err).Msg("Note not found")
		return utils.NotFound(c, "Note not found")
	}

	// fetch all the task for that particular note
	var tasks []models.Task
	if err := database.DB.Where("note_id=? AND user_id=?", noteID, userID).Find(&tasks).Error; err != nil {
		log.Error().Err(err).Msg("Failed to fetch notes")
		return utils.InternalError(c, "Failed to fetch tasks")
	}

	// send response
	return utils.Success(c, fiber.Map{"tasks": tasks})
}

// create tasks function
func CreateTask(c *fiber.Ctx) error {

	//get the userID from the JWT
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	// get note id from params
	noteID, err := uuid.Parse(c.Params("note_id"))
	if err != nil {
		return utils.BadRequest(c, "Invalid note id")
	}

	// check if note exists
	var note models.Note
	if err := database.DB.Where("id=? AND user_id=?", noteID, userID).First(&note).Error; err != nil {
		log.Error().Err(err).Msg("Note not found")
		return utils.NotFound(c, "Note not found")
	}

	//parse req body
	var req TaskRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid req body ")
	}

	// sanitze the req
	req.Title = taskSanitizer.Sanitize(strings.TrimSpace(req.Title))
	if req.Title == "" || len(req.Title) > 255 {
		return utils.BadRequest(c, "Title is required and must under 255 characters")
	}

	if req.Priority != "" && req.Priority != "high" {
		return utils.BadRequest(c, "Priority must be high or empty")
	}

	// create task
	task := models.Task{
		ID:       uuid.New(),
		NoteID:   noteID,
		UserID:   userID,
		Title:    req.Title,
		Status:   string(models.StatusPending),
		Priority: req.Priority,
	}

	// save to database
	if err := database.DB.Create(&task).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create task")
		return utils.InternalError(c, "Failed to create task")
	}

	// return response
	return utils.Created(c, task)
}

// update tasks function
func UpdateTask(c *fiber.Ctx) error {

	//get the userID from the JWT
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	// get task id from params
	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "Invalid task id")
	}

	// parse request body
	var req TaskRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid req body ")
	}

	// find task
	var task models.Task
	if err := database.DB.Where("id=? AND user_id=?", taskID, userID).First(&task).Error; err != nil {
		log.Error().Err(err).Msg("Failed to update task")
		return utils.NotFound(c, "Task not found")
	}

	// update status
	if req.Status != "" {
		if req.Status != string(models.StatusPending) && req.Status != string(models.StatusCompleted) {
			return utils.BadRequest(c, "Invalid status")
		}
		task.Status = req.Status
	}

	// update priority
	if req.Priority != "" {
		if req.Priority != "" && req.Priority != "high" {
			return utils.BadRequest(c, "Invalid priority")
		}
		task.Priority = req.Priority
	}

	if req.Priority == "" || req.Priority == " " {
		task.Priority = " "
	}

	// save updated task
	if err := database.DB.Save(&task).Error; err != nil {
		log.Error().Err(err).Msg("Failed to update task")
		return utils.InternalError(c, "Failed to update task")
	}

	// return response
	return utils.Success(c, task)
}

// delete tasks function
func DeleteTask(c *fiber.Ctx) error {

	//get the userID from the JWT
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	// get task id from params
	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "Invalid task id")
	}

	// find task
	var task models.Task
	if err := database.DB.Where("id=? AND user_id=?", taskID, userID).First(&task).Error; err != nil {
		log.Error().Err(err).Msg("Failed to update task")
		return utils.NotFound(c, "Task not found")
	}

	// delete task
	if err := database.DB.Delete(&task).Error; err != nil {
		log.Error().Err(err).Msg("Failed to delete task")
		return utils.InternalError(c, "Failed to delete task")
	}

	// return response
	return utils.Success(c, fiber.Map{"message": "Task deleted successfully"})
}
