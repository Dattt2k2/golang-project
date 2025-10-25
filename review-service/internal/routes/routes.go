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

	v1 := r.Group("/v1")
	{
		products := v1.Group("/products")
		{
			products.POST("/:product_id/reviews", h.CreateReview())
			products.GET("/:product_id/reviews", h.ListReviews())
		}
		
		reviews := v1.Group("/reviews")
		{
			reviews.GET("/:review_id", h.GetReview())
		}
	}
}
