package database

import (
	"context"
	"os"
	"time"
    "log"

	// "github.com/Dattt2k2/golang-project/auth-service/logger"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database
var Client *mongo.Client = DBinstance()

func DBinstance() *mongo.Client {
    // Load .env file
    err := godotenv.Load("./auth-service/.env")
    if err != nil {
        log.Println("Warning: Error loading .env file:", err)
        // logger.Err("Error loading .env file", err)
    }

    // Get MongoDB URL from environment variable
    MongoDB := os.Getenv("MONGODB_URL")
    if MongoDB == "" {
        log.Fatal("MONGODB_URL environment variable not set")
        // logger.Err("MONGODB_URL environment variable not set", nil)
    }

    // Set client options
    clientOptions := options.Client().ApplyURI(MongoDB)

    // Connect to MongoDB
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal("Failed to connect to MongoDB:", err)
        // logger.Err("Failed to connect to MongoDB", err)
    }

    // Ping the database
    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal("Failed to ping MongoDB:", err)
        // logger.Err("Failed to ping MongoDB", err)
    }

    log.Println("Connected to MongoDB")
    // logger.Info("Connected to MongoDB")

    return client
}



func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
    database := os.Getenv("MONGODB_DATABASE")
    if database == "" {
        log.Fatal("MONGODB_DATABASE environment variable not set")
        // logger.Err("MONGODB_DATABASE environment variable not set", nil)
    }
    return client.Database(database).Collection(collectionName)
}