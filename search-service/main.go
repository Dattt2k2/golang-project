package main

import (
	"log"
	"os"

	"github.com/Dattt2k2/golang-project/search-service/controller"
	"github.com/Dattt2k2/golang-project/search-service/database"
	"github.com/Dattt2k2/golang-project/search-service/repository"
	"github.com/Dattt2k2/golang-project/search-service/routes"
	"github.com/Dattt2k2/golang-project/search-service/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)


func main() {
	_ = godotenv.Load(".env")

	database.InitElasticsearch()

	repo := repository.NewSearchRepository()
	svc := service.NewSearchService(repo)
	ctrl := controller.NewSearchController(svc)

	router := gin.Default()
	routes.SearchRoutes(router, ctrl)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	log.Printf("Search service running on :%s", port)
	router.Run(":" + port)
}