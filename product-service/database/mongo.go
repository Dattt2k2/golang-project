
package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func DBinstance() *mongo.Client {
    // Load .env file
    err := godotenv.Load("github.com/Dattt2k2/golang-project/product-service/.env")
    if err != nil {
        log.Println("Warning: Error loading .env file:", err)
    }

    // Get MongoDB URL from environment variable
    MongoDB := os.Getenv("MONGODB_URL")
    if MongoDB == "" {
        log.Fatal("MONGODB_URL environment variable not set")
    }

    // Set client options
    clientOptions := options.Client().ApplyURI(MongoDB)

    // Connect to MongoDB
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal("Failed to connect to MongoDB:", err)
    }

    // Ping the database
    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal("Failed to ping MongoDB:", err)
    }

    fmt.Println("Connected to MongoDB")

    return client
}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
    database := os.Getenv("MONGODB_DATABASE")
    if database == "" {
        log.Fatal("MONGODB_DATABASE environment variable not set")
    }
    return client.Database(database).Collection(collectionName)
}