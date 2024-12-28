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

	err := godotenv.Load("github.com/Dattt2k2/golang-project/order-service/.env")

	if err != nil{
		log.Println("Warning: Error loading .env file")
	}

	MongoDB := os.Getenv("MONGODB_URL")
	if MongoDB == ""{
		log.Fatal("MONGODB_URL enviroment varialble not set")
	}

	clientOptions := options.Client().ApplyURI(MongoDB)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil{
		log.Fatal("Failed to connect to MONGODB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil{
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	fmt.Printf("Connected to MongoDB")

	return client
}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) * mongo.Collection{
	database := os.Getenv("MONGODB_DATABASE")
	if database == ""{
		log.Fatalf("MONGODB_DATABASE enviroment varialbe not set yet")

	}
	return client.Database(database).Collection(collectionName)
}