package main

import (
	"os"

	"github.com/Dattt2k2/golang-project/search-service/log"
	"github.com/Dattt2k2/golang-project/search-service/controller"
	"github.com/Dattt2k2/golang-project/search-service/database"
	"github.com/Dattt2k2/golang-project/search-service/kafka"
	"github.com/Dattt2k2/golang-project/search-service/repository"
	"github.com/Dattt2k2/golang-project/search-service/routes"
	"github.com/Dattt2k2/golang-project/search-service/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)


func main() {

	logger.InitLogger()
	defer logger.Sync()
	
	_ = godotenv.Load(".env")

	database.InitElasticsearch()

	repo := repository.NewSearchRepository()
	svc := service.NewSearchService(repo)
	ctrl := controller.NewSearchController(svc)

	router := gin.Default()
	routes.SearchRoutes(router, ctrl)

	kafkaHost := os.Getenv("KAFKA_URL")
	brokers := []string{kafkaHost}
	if kafkaHost == ""{
		brokers = []string{"localhost:9092"}
	}

	go kafka.InitProductEventConsumer(svc, brokers)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	logger.Logger.Info("Search service running on :%s", port)
	router.Run(":" + port)
}