package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"search-service/controller"
	"search-service/database"
	"search-service/kafka"
	"search-service/log"
	"search-service/repository"
	"search-service/routes"
	"search-service/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func waitForElasticsearch(url string, retries int, delay time.Duration) error {
	for i := 0; i < retries; i++ {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			return nil
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("Elasticsearch not available at %s", url)
}



func main() {

	logger.InitLogger()
	defer logger.Sync()
	
	_ = godotenv.Load(".env")

	

	repo := repository.NewSearchRepository()
	svc := service.NewSearchService(repo)
	ctrl := controller.NewSearchController(svc)

	router := gin.Default()
	routes.SearchRoutes(router, ctrl)

	kafkaHost := os.Getenv("KAFKA_URL")
	brokers := []string{kafkaHost}
	if kafkaHost == ""{
		brokers = []string{"kafka:9092"}
	}


	go kafka.InitProductEventConsumer(svc, brokers)

	err := waitForElasticsearch("http://elasticsearch:9200", 10, 3*time.Second)
	if err != nil {
		logger.Err("Elasticsearch not available: %v", err)
	}

	database.InitElasticsearch()

	err = svc.SyncProductFromProductService()
	if err != nil {
		logger.Err("Failed to sync products from product service: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	router.Run(":" + port)
}