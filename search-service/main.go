package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Dattt2k2/golang-project/search-service/controller"
	"github.com/Dattt2k2/golang-project/search-service/database"
	"github.com/Dattt2k2/golang-project/search-service/kafka"
	"github.com/Dattt2k2/golang-project/search-service/log"
	"github.com/Dattt2k2/golang-project/search-service/repository"
	"github.com/Dattt2k2/golang-project/search-service/routes"
	"github.com/Dattt2k2/golang-project/search-service/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func waitForElasticsearch(url string, retries int, delay time.Duration) error {
	for i := 0; i < retries; i++ {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			return nil
		}
		logger.Info("Waiting for Elasticsearch...")
		time.Sleep(delay)
	}
	return fmt.Errorf("Elasticsearch not available at %s", url)
}



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

	err := waitForElasticsearch("http://elasticsearch:9200", 5, 2*time.Second)
	if err != nil {
		logger.Err("Elasticsearch not available: %v", err)
	}

	err = svc.SyncProductFromProductService()
	if err != nil {
		logger.Err("Failed to sync products from product service: %v", err)
	}else{
		logger.Info("Successfully synced products from product service")
	}

	go kafka.InitProductEventConsumer(svc, brokers)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	logger.Logger.Info("Search service running on :%s", port)
	router.Run(":" + port)
}