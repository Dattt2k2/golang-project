package routes

import (
	"net/http"
	"review-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *handlers.ReviewHandler) {
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "review-service"})
	})

		products := r.Group("/products")
		{
			products.POST("create-reviews/:product_id", h.CreateReview())
			products.GET("/reviews/:product_id", h.ListReviews())
		}

		reviews := r.Group("/reviews")
		{
			reviews.GET("/:review_id", h.GetReview())
		}
}
