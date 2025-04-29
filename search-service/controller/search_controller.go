package controller

import (
	"net/http"

	"github.com/Dattt2k2/golang-project/search-service/service"
	"github.com/gin-gonic/gin"
)


type SearchController struct {
	service service.SearchService
}

func NewSearchController(service service.SearchService) *SearchController {
	return &SearchController {
		service: service,
	}
}


func (ctrl *SearchController) BasicSearch(query string) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("query")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query parameter"})
			return
		}
		
		products, err := ctrl.service.BasicSearch(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform search"})
			return
		}
		c.JSON(http.StatusOK, products)
	}
}


func (ctrl *SearchController) AdvancedSearch(query string, filters map[string]interface{}) gin.HandlerFunc {
	return func (c *gin.Context) {
		query := c.Query("query")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query parameter"})
			return
		}

		if price := c.Query("price"); price != "" {
			filters["price"] = map[string]interface{}{"lte": price}
		}
		if category := c.Query("category"); category != "" {
			filters["category"] = category
		}
		
		products, err := ctrl.service.AdvancedSearch(query, filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform search"})
			return
		}
		c.JSON(http.StatusOK, products)
	}
}



