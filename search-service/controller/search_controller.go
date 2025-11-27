package controller

import (
	"net/http"
	"strconv"
	"strings"

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
	}
}

func (ctrl *SearchController) AdvancedSearch() gin.HandlerFunc {
    return func(c *gin.Context) {
         q := c.DefaultQuery("q", "")
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
        if page < 1 { page = 1 }
        if limit < 1 { limit = 10 }
        from := (page - 1) * limit

        // parse raw sort params (strings)
        sortBy := strings.ToLower(c.DefaultQuery("sortBy", "created_at"))
        sortOrder := strings.ToLower(c.DefaultQuery("sortOrder", "desc"))
        if sortOrder != "asc" && sortOrder != "desc" { sortOrder = "desc" }

        // map sortBy string to int code (adjust based on service convention)
        var sortByCode int
        switch sortBy {
        case "name":
            sortByCode = 1
        case "price":
            sortByCode = 2
        case "rating":
            sortByCode = 3
        case "reviews", "review_count":
            sortByCode = 4
        default:
            sortByCode = 0 // created_at
        }

        // map sortOrder string to int (e.g., 1 = asc, -1 = desc)
        sortOrderCode := -1
        if sortOrder == "asc" {
            sortOrderCode = 1
        }

        filters := make(map[string]interface{})
        category := c.Query("category")
        if category != "" {
            filters["category"] = category
        }

        minPrice := c.Query("minPrice")
        if minPrice != "" {
            filters["price_min"] = minPrice
        }
        maxPrice := c.Query("maxPrice")
        if maxPrice != "" {
            filters["price_max"] = maxPrice
        }

        
        fromStr := strconv.Itoa(from)
        limitStr := strconv.Itoa(limit)

        response, err := ctrl.service.AdvancedSearch(q, filters, sortByCode, sortOrderCode, fromStr, limitStr) 
        if err != nil {
            logger.Err("Failed to perform advanced search", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, response)
    }
}
