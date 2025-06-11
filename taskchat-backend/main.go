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

// package main

// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"taskchat/config"
// 	"taskchat/database"
// 	"taskchat/handlers"
// 	"taskchat/middleware"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/gofiber/fiber/v2/middleware/cors"
// 	"github.com/gofiber/fiber/v2/middleware/limiter"
// 	"github.com/gofiber/fiber/v2/middleware/logger"
// 	"github.com/joho/godotenv"
// 	"github.com/rs/zerolog"
// 	zlog "github.com/rs/zerolog/log"
// )

// // App holds the Fiber app and database connection
// type App struct {
// 	Fiber *fiber.App // The web server
// 	DB    *database.DB // The database connection
// }

// // NewApp sets up the app with database and routes
// func NewApp() (*App, error) {
// 	// Load environment variables from .env file
// 	if err := godotenv.Load(); err != nil {
// 		return nil, fmt.Errorf("cannot load .env file: %v", err)
// 	}

// 	// Check if required environment variables are set
// 	if err := config.ValidateEnvVars(); err != nil {
// 		return nil, fmt.Errorf("missing environment variables: %v", err)
// 	}

// 	// Set up logger to print clear logs
// 	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
// 	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})

// 	// Connect to the database
// 	db, err := database.InitDB()
// 	if err != nil {
// 		return nil, fmt.Errorf("cannot connect to database: %v", err)
// 	}

// 	// Create a new Fiber app
// 	fiberApp := fiber.New(fiber.Config{
// 		ErrorHandler: handlers.ErrorHandler, // Custom error handling
// 	})

// 	// Add middleware for logging, CORS, and rate limiting
// 	fiberApp.Use(logger.New()) // Logs all requests
// 	fiberApp.Use(cors.New(cors.Config{
// 		AllowOrigins: os.Getenv("CORS_ALLOW_ORIGINS"), // Allow specific frontend URLs
// 		AllowMethods: "GET,POST,PUT,DELETE",          // Allowed HTTP methods
// 	}))
// 	fiberApp.Use(limiter.New(limiter.Config{
// 		Max:        5,           // Allow 5 requests per minute
// 		Expiration: 60,          // Per minute
// 		KeyGenerator: func(c *fiber.Ctx) string {
// 			return c.IP() // Limit by client IP
// 		},
// 		LimitReached: func(c *fiber.Ctx) error {
// 			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
// 				"success": false,
// 				"error":   "Too many requests, try again later",
// 			})
// 		},
// 	}))

// 	// Define API routes
// 	fiberApp.Get("/api/health", func(c *fiber.Ctx) error {
// 		return c.JSON(fiber.Map{"status": "ok"}) // Health check endpoint
// 	})
// 	fiberApp.Post("/api/register", handlers.Register(db))
// 	fiberApp.Post("/api/login", handlers.Login(db))

// 	// Protected routes require JWT authentication
// 	protected := fiberApp.Group("/api", middleware.AuthMiddleware)
// 	protected.Get("/protected", func(c *fiber.Ctx) error {
// 		return c.JSON(fiber.Map{"message": "You accessed a protected route!"})
// 	})

// 	// Note routes
// 	notes := protected.Group("/notes")
// 	notes.Get("/", handlers.GetNotes(db))
// 	notes.Post("/", handlers.CreateNote(db))
// 	notes.Put("/:id", handlers.UpdateNote(db))
// 	notes.Delete("/:id", handlers.DeleteNote(db))

// 	// Task routes
// 	tasks := protected.Group("/notes/:note_id/tasks")
// 	tasks.Get("/", handlers.GetTasks(db))
// 	tasks.Post("/", handlers.CreateTask(db))
// 	tasks.Put("/tasks/:id", handlers.UpdateTask(db))
// 	tasks.Delete("/tasks/:id", handlers.DeleteTask(db))

// 	// Priority tasks route
// 	protected.Get("/priorities", handlers.GetPriorities(db))

// 	return &App{Fiber: fiberApp, DB: db}, nil
// }

// func main() {
// 	// Create the app
// 	app, err := NewApp()
// 	if err != nil {
// 		log.Fatalf("Failed to start app: %v", err)
// 	}

// 	// Start the server in a separate goroutine
// 	go func() {
// 		port := config.LoadConfig().Port
// 		if err := app.Fiber.Listen(":" + port); err != nil {
// 			zlog.Fatal().Err(err).Msg("Cannot start server")
// 		}
// 	}()

// 	// Wait for Ctrl+C to shut down gracefully
// 	c := make(chan os.Signal, 1)
// 	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
// 	<-c
// 	zlog.Info().Msg("Shutting down server...")
// 	if err := app.Fiber.Shutdown(); err != nil {
// 		zlog.Error().Err(err).Msg("Error shutting down server")
// 	}
// 	zlog.Info().Msg("Server stopped")
// }
