// package database

// import (
// 	"auth-service/logger"
// 	"fmt"
// 	"os"

// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// var DB *gorm.DB

// func InitDB() *gorm.DB {
// 	host := os.Getenv("POSTGRES_HOST")
// 	port := os.Getenv("POSTGRES_PORT")
// 	user := os.Getenv("POSTGRES_USER")
// 	password := os.Getenv("POSTGRES_PASSWORD")
// 	dbname := os.Getenv("POSTGRES_DB")
// 	sslmode := os.Getenv("POSTGRES_SSLMODE")

// 	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
// 		host, user, password, dbname, port, sslmode,
// 	)
// 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		logger.Logger.Fatal("Failed to connect to PostgreSQL", logger.ErrField(err))
// 	}
// 	return db
// }


package database

import (
    "fmt"
    "os"
    "time"

    "auth-service/logger"
    "go.uber.org/zap"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
    host := os.Getenv("POSTGRES_HOST")
    port := os.Getenv("POSTGRES_PORT")
    user := os.Getenv("POSTGRES_USER")
    pass := os.Getenv("POSTGRES_PASSWORD")
    dbname := os.Getenv("POSTGRES_DB")
    ssl := os.Getenv("POSTGRES_SSLMODE")

    // 1. Validate
    if host == "" || port == "" || user == "" || pass == "" || dbname == "" {
        logger.Logger.Fatal("Postgres env vars missing",
            zap.String("host", host),
            zap.String("port", port),
            zap.String("user", user),
            zap.String("dbname", dbname),
        )
    }

    // 2. Log trước khi connect
    logger.Logger.Info("Connecting to Postgres",
        zap.String("host", host),
        zap.String("port", port),
        zap.String("user", user),
        zap.String("dbname", dbname),
        zap.String("sslmode", ssl),
    )

    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        host, port, user, pass, dbname, ssl,
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        logger.Logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
    }

    // 3. Ping to verify
    sqlDB, _ := db.DB()
    sqlDB.SetConnMaxLifetime(time.Minute * 10)
    if err := sqlDB.Ping(); err != nil {
        logger.Logger.Fatal("Postgres ping failed", zap.Error(err))
    }

    DB = db
    return DB
}