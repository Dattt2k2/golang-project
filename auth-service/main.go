package main

import (
	// "context"
	"log"
	"os"
	"time"

	// "github.com/Dattt2k2/golang-project/auth-service/database"
	"github.com/Dattt2k2/golang-project/auth-service/database"
	"github.com/Dattt2k2/golang-project/auth-service/helpers"
	"github.com/Dattt2k2/golang-project/auth-service/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)


var userBloomFilter *helpers.BloomFilter

func initBloomFilter(userCollection *mongo.Collection){
	userBloomFilter = helpers.CreateOptimalUserBloomFilter(100000)

	if err := userBloomFilter.Init(userCollection); err != nil{
		log.Printf("Error initializing bloom filter: %v", err)
	} else{
		log.Println("Bloom filter initialized")
	}

	go func(){
		for {
			time.Sleep(24 *time.Hour)

			if err := userBloomFilter.Init(userCollection); err != nil{
				log.Printf("Error update bloom filter: %v", err)
			} else{
				log.Println("Bloom filter updated")
			}
		}
	}()
}



func main(){
	err := godotenv.Load("./auth-service/.env")
    if err != nil {
        log.Println("Warning2 : Error loading .env file:", err)
    }
	mongodbURL := os.Getenv("MONGODB_URL")
	if mongodbURL == ""{
		log.Fatalf("MONGODB_URL is not set on .env file yet")
	}

	database.InitRedis()
	defer database.RedisClient.Close()
	log.Printf("Connected to Redis")

	userCollection := database.OpenCollection(database.Client, "user")
	initBloomFilter(userCollection)

	helpers.SetUserBloomFilter(userBloomFilter)

	// mongoClient := database.DBinstance()
    // defer mongoClient.Disconnect(context.Background())

	port := os.Getenv("PORT")
	if port == ""{
		port = "8081"
	}
	helpers.InitDotEnv()

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.Run(":"+ port)
	
}