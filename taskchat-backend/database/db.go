package database

import (
	"log"
	"os"
	"taskchat/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("Error connecting database from .env")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Error connecting database")
	}

	DB = db
	log.Println("Database connected")

	err = db.AutoMigrate(&models.User{}, &models.Note{}, &models.Task{})
	if err != nil {
		log.Fatal("Error migrating database")
	}

	log.Println("migration connected")
}
