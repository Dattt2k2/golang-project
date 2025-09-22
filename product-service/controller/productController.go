package controllers

import (
	"context"
	"log"
	"net/http"

	"strconv"
	"time"

	"product-service/models"
	"product-service/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductController struct {
	service   service.ProductService
	s3Service *service.S3Service
}

func NewProductController(service service.ProductService, s3Service service.S3Service) *ProductController {
	return &ProductController{
		service:   service,
		s3Service: &s3Service,
	}
}

func (ctrl *ProductController) AddProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()
		// CheckSellerRole(c)
		if c.IsAborted() {
			return
		}

		userID := c.GetHeader("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found"})
			return
		}
		var req models.CreateProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
			return
		}
		name := req.Name
		description := req.Description
		quantity := req.Quantity
		price := req.Price
		category := req.Category
		imagePath := req.ImagePath

		product := models.Product{
			Name:        name,
			Category:    category,
			Description: description,
			Price:       price,
			Quantity:    quantity,
			ImagePath:   imagePath,
			UserID:      userID,
		}

		if err := ctrl.service.AddProduct(ctx, product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Product added successfully"})
	}
}

func (ctrl *ProductController) EditProduct() gin.HandlerFunc {
	return func(c *gin.Context) {

		// CheckSellerRole(c)
		if c.IsAborted() {
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID not found"})
			return
		}

		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Product ID"})
			return
		}

		var req models.UpdateProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
			return 
		}

		update := make(map[string]interface{})
		if req.Name != nil {
			update["name"] = *req.Name
		}
		if req.ImagePath != nil {
			update["image_path"] = *req.ImagePath
		}
		if req.Category != nil {
			update["category"] = *req.Category
		}
		if req.Description != nil {
			update["description"] = *req.Description
		}
		if req.Quantity != nil {
			update["quantity"] = *req.Quantity
		}
		if req.Price != nil {
			update["price"] = *req.Price
		}

		if len(update) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		if err := ctrl.service.EditProduct(ctx, id, update); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
	}
}

func (ctrl *ProductController) DeleteProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		// CheckSellerRole(c)
		if c.IsAborted() {
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		userID, exits := c.Get("uid")
		if !exits {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found"})
			return
		} 
		userIDStr, ok := userID.(string)
		if !ok || userIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
			return
		}
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID not found"})
			return
		}

		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Product ID"})
			return
		}

		if err := ctrl.service.DeleteProduct(ctx, id, userIDStr); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
	}
}

func (ctrl *ProductController) GetAllProducts() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Starting GetAllProducts handler")

		var ctx, cancel = context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		// Parse pagination parameters
		page, err := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
		if err != nil || page < 1 {
			log.Printf("Invalid page parameter, using default: %v", err)
			page = 1
		}

		limit, err := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
		if err != nil || limit < 1 {
			log.Printf("Invalid limit parameter, using default: %v", err)
			limit = 10
		}

		log.Printf("Pagination: page=%d, limit=%d", page, limit)

		// Call service layer
		products, total, pages, hasNext, hasPrev, cached, err := ctrl.service.GetAllProducts(ctx, page, limit)
		if err != nil {
			log.Printf("Error fetching products: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		// Debug info
		log.Printf("Found %d products (total: %d)", len(products), total)

		response := gin.H{
			"data":     products,
			"total":    total,
			"page":     page,
			"pages":    pages,
			"has_next": hasNext,
			"has_prev": hasPrev,
			"cached":   cached,
		}

		log.Printf("Sending response with %d products (cached: %v)", len(products), cached)
		c.JSON(http.StatusOK, response)
	}
}

// func (ctrl *ProductController) GetProductByName() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		name := c.Query("name")
// 		if name == "" {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Name query parameter is required"})
// 			return
// 		}
// 		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
// 		defer cancel()
// 		products, err := ctrl.service.GetProductByName(ctx, name)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		if len(products) == 0 {
// 			c.JSON(http.StatusNotFound, gin.H{"message": "No product found"})
// 			return
// 		}
// 		c.JSON(http.StatusOK, products)
// 	}
// }

type StockUpdateItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// Update product stock in the database
// isRestock: true for restock, false for sale
func (ctrl *ProductController) UpdateProductStock(ctx context.Context, items []StockUpdateItem, isRestock bool) error {
	for _, item := range items {

		quantity := item.Quantity
		if !isRestock {
			quantity = -item.Quantity
		}

		// Call UpdateProductStock with proper parameters (product ID and quantity)
		err := ctrl.service.UpdateProductStock(ctx, item.ProductID, quantity)
		if err != nil {
			return err
		}
	}
	return nil

}

func (ctrl *ProductController) GetBestSellingProducts() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()
		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 10
		}

		products, err := ctrl.service.GetBestSellingProducts(ctx, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  products,
			"count": len(products),
		})
	}
}

func (ctrl *ProductController) IncrementSoldCount(ctx context.Context, productID string, quantity int) error {
	return ctrl.service.IncrementSoldCount(ctx, productID, quantity)
}

func (ctrl *ProductController) DecrementSoldCount(ctx context.Context, productID string, quantity int) error {
	return ctrl.service.DecrementSoldCount(ctx, productID, quantity)
}

// // CreateProduct - Workflow 1: Tạo product với image_path có sẵn (từ presigned URL)
// func (pc *ProductController) CreateProduct(c *gin.Context) {
// 	var req models.CreateProductRequest

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error":   "Invalid request data",
// 			"details": err.Error(),
// 		})
// 		return
// 	}

// 	// Validate image_path if provided
// 	if req.ImagePath != "" {
// 		// Optional: Validate if URL is accessible or from your S3 bucket
// 		log.Printf("Product will be created with image: %s", req.ImagePath)
// 	}

// 	// Convert request to model
// 	product := models.Product{
// 		ID:          string,
// 		Name:        req.Name,
// 		ImagePath:   req.ImagePath, // Có thể empty hoặc có URL
// 		Category:    req.Category,
// 		Description: req.Description,
// 		Quantity:    req.Quantity,
// 		Price:       req.Price,
// 		SoldCount:   0,
// 		Created_at:  time.Now(),
// 		Updated_at:  time.Now(),
// 		// UserID sẽ được set từ JWT token
// 	}

// 	// Save to database
// 	err := pc.service.AddProduct(c.Request.Context(), product)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error":   "Failed to create product",
// 			"details": err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"success": true,
// 		"message": "Product created successfully",
// 	})
// }

