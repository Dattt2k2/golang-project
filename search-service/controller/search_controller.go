package controller

import (
	"net/http"

	logger "search-service/log"
	"search-service/service"

	"github.com/gin-gonic/gin"
)

type SearchController struct {
	service service.SearchService
}

func NewSearchController(service service.SearchService) *SearchController {
	return &SearchController{
		service: service,
	}
}

func (ctrl *SearchController) BasicSearch(query string) gin.HandlerFunc {
	return func(c *gin.Context) {
	   logger.Logger.Infof("RawQuery: %s", c.Request.URL.RawQuery)
		query := c.Query("query")
		if query == "" {
			query = c.Query("q")
		}
		if query == "" {
			logger.Err("Missing query parameter", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query parameter"})
			return
		}
		products, err := ctrl.service.BasicSearch(query)
		if err != nil {
			logger.Err("Failed to perform search", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform search"})
			return
		}
		c.JSON(http.StatusOK, products)
		logger.Logger.Infof("Search results for query '%s': %v", query, products)
	}
}

func (ctrl *SearchController) AdvancedSearch(query string, filters map[string]interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
	   logger.Logger.Infof("RawQuery: %s", c.Request.URL.RawQuery)
		query := c.Query("query")
		if query == "" {
			query = c.Query("q")
		}
		if query == "" {
			logger.Err("Missing query parameter", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query parameter"})
			return
		}
		if price := c.Query("price"); price != "" {
			logger.Logger.Infof("Price filter applied: %s", price)
			filters["price"] = map[string]interface{}{"lte": price}
		}
		if category := c.Query("category"); category != "" {
			logger.Logger.Infof("Category filter applied: %s", category)
			filters["category"] = category
		}
		products, err := ctrl.service.AdvancedSearch(query, filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform search"})
			return
		}
		c.JSON(http.StatusOK, products)
		logger.Logger.Infof("Search results for query '%s' with filters %v: %v", query, filters, products)
	}
}
