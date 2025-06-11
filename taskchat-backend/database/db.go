package database

import (
	"fmt"
	"os"
	model "taskchat/models"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB holds the database connection
type DB struct {
	Conn *gorm.DB
}

// InitDB connects to the database
func InitDB() (*DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("cannot get database instance: %v", err)
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Info().Msg("Database connected")

	if err := db.AutoMigrate(&model.User{}, &model.Note{}, &model.Task{}); err != nil {
		return nil, fmt.Errorf("cannot migrate database: %v", err)
	}

	log.Info().Msg("Database migrated")
	return &DB{Conn: db}, nil
}
