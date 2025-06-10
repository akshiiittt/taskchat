package main

import (
	"log"
	"os"
	"taskchat/database"
	"taskchat/handlers"
	"taskchat/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	database.InitDB()

	app := fiber.New()

	app.Use(logger.New())

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Post("/api/register", handlers.Register)
	app.Post("/api/login", handlers.Login)

	app.Get("/api/protected", middleware.AuthMiddleware, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Protected route accessed"})
	})

	notes := app.Group("/api/notes", middleware.AuthMiddleware)
	notes.Get("/", handlers.GetNotes)
	notes.Post("/", handlers.CreateNote)
	notes.Put("/:id", handlers.UpdateNote)
	notes.Delete("/:id", handlers.DeleteNote)

	tasks := app.Group("/api", middleware.AuthMiddleware)
	tasks.Get("/notes/:note_id/tasks", handlers.GetTasks)
	tasks.Post("/notes/:note_id/tasks", handlers.CreateTask)
	tasks.Put("/tasks/:id", handlers.UpdateTask)
	tasks.Delete("/tasks/:id", handlers.DeleteTask)

	app.Get("/api/priorities", middleware.AuthMiddleware, handlers.GetPriorities)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	err := app.Listen(":" + port)
	if err != nil {
		log.Fatal("Error starting server")
	}
}
