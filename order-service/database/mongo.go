package database

import (
	"context"
	"fmt"
	"os"
	"time"

	logger "github.com/Dattt2k2/golang-project/order-service/log"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)



func DBinstance() *mongo.Client {

	err := godotenv.Load("github.com/Dattt2k2/golang-project/order-service/.env")

	if err != nil{
		logger.Err("Error loading .env file", err)
	}

	MongoDB := os.Getenv("MONGODB_URL")
	if MongoDB == ""{
		logger.Err("MONGODB_URL enviroment varialble not set", nil)
	}

	clientOptions := options.Client().ApplyURI(MongoDB)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil{
		logger.Err("Failed to connect to MONGODB", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil{
		logger.Err("Failed to ping MONGODB", err)
	}

	fmt.Printf("Connected to MongoDB")

	return client
}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) * mongo.Collection{
	database := os.Getenv("MONGODB_DATABASE")
	if database == ""{
		logger.Err("MONGODB_DATABASE enviroment varialble not set", nil)

	}
	return client.Database(database).Collection(collectionName)
}