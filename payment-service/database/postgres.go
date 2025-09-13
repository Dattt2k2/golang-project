package database

import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
    "os"
)

func NewPostgres() (*gorm.DB, error) {
    dsn := os.Getenv("DATABASE_URL")
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    return db, nil
}