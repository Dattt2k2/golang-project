package controllers

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Dattt2k2/golang-project/product-service/models"
	"github.com/Dattt2k2/golang-project/product-service/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// func CheckSellerRole(c *gin.Context) {
// 	userRole := c.GetHeader("user_type")
// 	if userRole != "SELLER" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "You don't have permission"})
// 		c.Abort()
// 		return
// 	}
// 	c.Next()
// }

func saveImageToFileSystem(c *gin.Context, file *multipart.FileHeader) (string, error) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Error getting working directory: %v", err)
	}
	log.Printf("Current working directory: %s", wd)

	// Create absolute paths
	possibleDirs := []string{
		filepath.Join(wd, "uploads", "images"),
		filepath.Join(wd, "product-service", "uploads", "images"),
	}

	var saveDir string
	for _, dir := range possibleDirs {
		err := os.MkdirAll(dir, os.ModePerm)
		if err == nil {
			saveDir = dir
			log.Printf("Successfully created directory: %s", saveDir)
			break
		}
		log.Printf("Failed to create directory %s: %v", dir, err)
	}

	if saveDir == "" {
		return "", fmt.Errorf("Failed to create any image directory")
	}

	// Create a unique filename
	imageFileName := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)
	imagePath := filepath.Join(saveDir, imageFileName)

	log.Printf("Saving file to: %s", imagePath)

	// Save the file
	if err := c.SaveUploadedFile(file, imagePath); err != nil {
		return "", fmt.Errorf("Failed to save image: %v", err)
	}

	log.Printf("Successfully saved image to: %s", imagePath)

	// Return just the filename
	return imageFileName, nil
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

		userObjID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		name := c.PostForm("name")
		description := c.PostForm("description")
		quantityStr := c.PostForm("quantity")
		priceStr := c.PostForm("price")
		category := c.PostForm("category")

		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quantity"})
			return
		}
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price"})
			return
		}

		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
			return
		}

		imagePath, err := saveImageToFileSystem(c, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}

		product := models.Product{
			Name:        name,
			Category:    category,
			Description: description,
			Price:       price,
			Quantity:    quantity,
			ImagePath:   imagePath,
			UserID:      userObjID,
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
		productID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		name := c.PostForm("name")
		description := c.PostForm("description")
		priceStr := c.PostForm("price")
		quantityStr := c.PostForm("quantity")

		// update := bson.M{"updated_at": time.Now()}
		update := bson.M{}
		if name != "" {
			update["name"] = name
		}

		if description != "" {
			update["description"] = description
		}
		if priceStr != "" {
			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price format"})
				return
			}
			update["price"] = price
		}
		if quantityStr != "" {
			quantity, err := strconv.Atoi(quantityStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quantity format"})
				return
			}
			update["quantity"] = quantity
		}

		file, err := c.FormFile("image")
		if err == nil {
			imagePath, err := saveImageToFileSystem(c, file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			update["image_path"] = imagePath
		}

		if err := ctrl.service.EditProduct(ctx, productID, update); err != nil {
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

		userID, err := primitive.ObjectIDFromHex(c.GetHeader("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		}
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID not found"})
			return
		}
		productID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Product ID"})
			return
		}

		if err := ctrl.service.DeleteProduct(ctx, productID, userID); err != nil {
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

func (ctrl *ProductController) GetProductByName() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name query parameter is required"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()
		products, err := ctrl.service.GetProductByName(ctx, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if len(products) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "No product found"})
			return
		}
		c.JSON(http.StatusOK, products)
	}
}

type StockUpdateItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// Update product stock in the database
// isRestock: true for restock, false for sale
func (ctrl *ProductController) UpdateProductStock(ctx context.Context, items []StockUpdateItem, isRestock bool) error {
	for _, item := range items {
		objID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			return err
		}

		quantity := item.Quantity
		if !isRestock {
			quantity = -item.Quantity
		}

		// Call UpdateProductStock with proper parameters (product ID and quantity)
		err = ctrl.service.UpdateProductStock(ctx, objID, quantity)
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

// CreateProduct - Workflow 1: Tạo product với image_path có sẵn (từ presigned URL)
func (pc *ProductController) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Validate image_path if provided
	if req.ImagePath != "" {
		// Optional: Validate if URL is accessible or from your S3 bucket
		log.Printf("Product will be created with image: %s", req.ImagePath)
	}

	// Convert request to model
	product := models.Product{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		ImagePath:   req.ImagePath, // Có thể empty hoặc có URL
		Category:    req.Category,
		Description: req.Description,
		Quantity:    req.Quantity,
		Price:       req.Price,
		SoldCount:   0,
		Created_at:  time.Now(),
		Updated_at:  time.Now(),
		// UserID sẽ được set từ JWT token
	}

	// Save to database
	err := pc.service.AddProduct(c.Request.Context(), product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create product",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Product created successfully",
	})
}

// CreateProductWithImage - Workflow 2: Tạo product và upload ảnh cùng lúc
func (pc *ProductController) CreateProductWithImage(c *gin.Context) {
	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	// Get product data from form
	var req models.CreateProductWithImageRequest
	req.Name = c.PostForm("name")
	req.Category = c.PostForm("category")
	req.Description = c.PostForm("description")

	if req.Name == "" || req.Category == "" || req.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields: name, category, description",
		})
		return
	}

	// Parse numeric fields
	quantity, err := strconv.Atoi(c.PostForm("quantity"))
	if err != nil || quantity < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid quantity",
		})
		return
	}
	req.Quantity = quantity

	price, err := strconv.ParseFloat(c.PostForm("price"), 64)
	if err != nil || price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid price",
		})
		return
	}
	req.Price = price

	// Handle file upload
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Image file is required",
		})
		return
	}
	defer file.Close()

	// Upload image to S3
	imageURL, err := pc.s3Service.UploadFile(file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to upload image",
			"details": err.Error(),
		})
		return
	}

	// Create product with uploaded image URL
	product := models.Product{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		ImagePath:   imageURL, // URL từ S3
		Category:    req.Category,
		Description: req.Description,
		Quantity:    req.Quantity,
		Price:       req.Price,
		SoldCount:   0,
		Created_at:  time.Now(),
		Updated_at:  time.Now(),
	}

	// Save to database
	err = pc.service.AddProduct(c.Request.Context(), product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create product",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Product created successfully with image",
	})
}
