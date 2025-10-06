package main

import (
    "fmt"
    "log"

    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    cfg "user-service/internal/config"
    "user-service/internal/events"
    "user-service/internal/handlers"
    "user-service/internal/models"
    "user-service/internal/repository"
    "user-service/internal/routes"
    "user-service/internal/services"
)

func main() {
    config := cfg.InitConfig()

    d := config.Database
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        d.Host, d.User, d.Password, d.Name, d.Port)

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("failed to connect database: %v", err)
    }

    // Auto migrate
    if err := db.AutoMigrate(&models.User{}); err != nil {
        log.Fatalf("failed to migrate database: %v", err)
    }

    if err := db.AutoMigrate(&models.UserAddress{}); err != nil {
        log.Fatalf("failed to migrate database: %v", err)
    }

    // Initialize layers
    userRepo := repository.NewUserRepository(db)

    // Kafka configuration (optional) - use env from config package only
    var pub events.EventPublisher
    kafkaBrokers := cfg.GetEnv("KAFKA_BROKERS", "")
    if kafkaBrokers != "" {
        brokers := cfg.SplitAndTrim(kafkaBrokers, ",")
        topic := cfg.GetEnv("KAFKA_TOPIC_USERS", "user.created")
        log.Printf("Kafka brokers=%v topic=%s", brokers, topic)
        pub = events.NewKafkaPublisher(brokers, topic)
        // start consumer (reads user.created)
        events.StartUserCreatedConsumer(brokers, topic, userRepo)
    } else {
        log.Println("KAFKA_BROKERS not set, using LoggingPublisher")
        pub = events.NewLoggingPublisher()
    }

    userService := services.NewUserService(userRepo, pub)
    userHandler := &handlers.UserHandler{UserService: userService}

    // Setup Gin
    r := gin.Default()

    routes.Register(r, userHandler)

    addr := fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port)
    log.Printf("Starting server on %s\n", addr)
    if err := r.Run(addr); err != nil {
        log.Fatalf("server exited with: %v", err)
    }
}