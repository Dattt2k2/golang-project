package database

import (
	"fmt"
	logger "order-service/log"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		logger.Logger.Fatal("Failed to get database instance", zap.Error(err))
	}

	// Configure connection pool to handle high concurrency
	sqlDB.SetMaxIdleConns(25)                  // Minimum number of idle connections
	sqlDB.SetMaxOpenConns(100)                 // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Maximum lifetime of a connection
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Maximum idle time for a connection

	logger.Logger.Info("Database connection pool configured",
		zap.Int("max_idle_conns", 25),
		zap.Int("max_open_conns", 100),
	)

	DB = db
	return DB
}
