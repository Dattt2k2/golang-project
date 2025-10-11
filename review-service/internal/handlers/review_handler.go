package handlers

import (
	"context"
	"net/http"
	"strconv"

	"review-service/internal/models"
	"review-service/internal/services"
	logger "review-service/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ReviewHandler struct {
	service services.ReviewService
}

func NewReviewHandler(service services.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (h *ReviewHandler) CreateReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.Review
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
			return
		}

		productID := c.Param("product_id")
		if productID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product_id required"})
			return
		}
		input.ProductID = productID

		ctx := context.Background()
		if err := h.service.Save(ctx, &input); err != nil {
			logger.Error("Failed to save review: ", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
			return
		}
		c.JSON(http.StatusCreated, input)
	}
}

func (h *ReviewHandler) ListReviews() gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("product_id")
		if productID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product_id required"})
			return
		}

		ctx := context.Background()
		limit := 50
		if q := c.Query("limit"); q != "" {
			if l, err := strconv.Atoi(q); err == nil && l > 0 {
				limit = l
			}
		}

		// added: read pagination key from query params
		lastKey := c.Query("lastKey")
		if lastKey == "" {
			lastKey = c.Query("last_key")
		}

		revs, nextKey, err := h.service.GetByProductID(ctx, productID, limit, lastKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"reviews": revs, "count": len(revs), "next_key": nextKey})
	}
}

func (h *ReviewHandler) GetReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		reviewID := c.Param("review_id")
		if reviewID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "review_id required"})
			return
		}

		ctx := context.Background()
		rev, err := h.service.GetByID(ctx, reviewID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
			return
		}
		c.JSON(http.StatusOK, rev)
	}
}
