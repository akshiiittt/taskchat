package database

import (
	"fmt"
	"time"

	"os"
	"taskchat/models"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("error connecting database from .env")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("error connecting database %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("cannot get the datbase instance %v", err)
	}

	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println(db, "*******************")
	fmt.Println(DB, "*******************")

	DB = db
	log.Info().Msg("Database connected")

	err = db.AutoMigrate(&models.User{}, &models.Note{}, &models.Task{})
	if err != nil {
		return fmt.Errorf("error migrating database %v", err)
	}

	log.Info().Msg("migration connected")

	return nil
}
