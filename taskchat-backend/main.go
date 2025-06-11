package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"taskchat/database"
	"taskchat/handlers"
	"taskchat/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// App holds the Fiber app and database
type App struct {
	Fiber *fiber.App
	DB    *database.DB
}

// NewApp initializes the app
func NewApp() (*App, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("cannot load .env: %v", err)
	}

	// Set up zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		return nil, fmt.Errorf("cannot initialize database: %v", err)
	}

	// Create Fiber app
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler: handlers.ErrorHandler,
	})

	// Add middleware
	fiberApp.Use(logger.New())
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("CORS_ALLOW_ORIGINS"),
		AllowMethods: "GET,POST,PUT,DELETE",
	}))
	fiberApp.Use(limiter.New(limiter.Config{
		Max:        5,
		Expiration: 60,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Too many requests",
			})
		},
	}))

	// Define routes
	fiberApp.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	fiberApp.Post("/api/register", handlers.Register(db))
	fiberApp.Post("/api/login", handlers.Login(db))

	protected := fiberApp.Group("/api", middleware.AuthMiddleware)
	protected.Get("/protected", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Protected route accessed"})
	})

	notes := protected.Group("/notes")
	notes.Get("/", handlers.GetNotes(db))
	notes.Post("/", handlers.CreateNote(db))
	notes.Put("/:id", handlers.UpdateNote(db))
	notes.Delete("/:id", handlers.DeleteNote(db))

	tasks := protected.Group("/notes/:note_id/tasks")
	tasks.Get("/", handlers.GetTasks(db))
	tasks.Post("/", handlers.CreateTask(db))
	tasks.Put("/:id", handlers.UpdateTask(db))
	tasks.Delete("/:id", handlers.DeleteTask(db))

	protected.Get("/priorities", handlers.GetPriorities(db))

	return &App{Fiber: fiberApp, DB: db}, nil
}

func main() {
	// Initialize app
	app, err := NewApp()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start app")
	}

	// Start server in a goroutine
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8081"
		}
		if err := app.Fiber.Listen(":" + port); err != nil {
			log.Fatal().Err(err).Msg("Error starting server")
		}
	}()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Info().Msg("Shutting down server...")
	if err := app.Fiber.Shutdown(); err != nil {
		log.Error().Err(err).Msg("Error shutting down server")
	}
	log.Info().Msg("Server stopped")
}
