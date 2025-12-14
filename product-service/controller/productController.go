package controllers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	logger "product-service/log"
	"product-service/models"
	"product-service/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductController struct {
	service   service.ProductService
	s3Service *service.S3Service
}

// removeDiacritics loại bỏ dấu tiếng Việt
func removeDiacritics(s string) string {
	// Map từng ký tự
	var result strings.Builder
	for _, r := range s {
		switch r {
		case 'à', 'á', 'ạ', 'ả', 'ã', 'â', 'ầ', 'ấ', 'ậ', 'ẩ', 'ẫ', 'ă', 'ằ', 'ắ', 'ặ', 'ẳ', 'ẵ':
			result.WriteRune('a')
		case 'À', 'Á', 'Ạ', 'Ả', 'Ã', 'Â', 'Ầ', 'Ấ', 'Ậ', 'Ẩ', 'Ẫ', 'Ă', 'Ằ', 'Ắ', 'Ặ', 'Ẳ', 'Ẵ':
			result.WriteRune('A')
		case 'è', 'é', 'ẹ', 'ẻ', 'ẽ', 'ê', 'ề', 'ế', 'ệ', 'ể', 'ễ':
			result.WriteRune('e')
		case 'È', 'É', 'Ẹ', 'Ẻ', 'Ẽ', 'Ê', 'Ề', 'Ế', 'Ệ', 'Ể', 'Ễ':
			result.WriteRune('E')
		case 'ì', 'í', 'ị', 'ỉ', 'ĩ':
			result.WriteRune('i')
		case 'Ì', 'Í', 'Ị', 'Ỉ', 'Ĩ':
			result.WriteRune('I')
		case 'ò', 'ó', 'ọ', 'ỏ', 'õ', 'ô', 'ồ', 'ố', 'ộ', 'ổ', 'ỗ', 'ơ', 'ờ', 'ớ', 'ợ', 'ở', 'ỡ':
			result.WriteRune('o')
		case 'Ò', 'Ó', 'Ọ', 'Ỏ', 'Õ', 'Ô', 'Ồ', 'Ố', 'Ộ', 'Ổ', 'Ỗ', 'Ơ', 'Ờ', 'Ớ', 'Ợ', 'Ở', 'Ỡ':
			result.WriteRune('O')
		case 'ù', 'ú', 'ụ', 'ủ', 'ũ', 'ư', 'ừ', 'ứ', 'ự', 'ử', 'ữ':
			result.WriteRune('u')
		case 'Ù', 'Ú', 'Ụ', 'Ủ', 'Ũ', 'Ư', 'Ừ', 'Ứ', 'Ự', 'Ử', 'Ữ':
			result.WriteRune('U')
		case 'ỳ', 'ý', 'ỵ', 'ỷ', 'ỹ':
			result.WriteRune('y')
		case 'Ỳ', 'Ý', 'Ỵ', 'Ỷ', 'Ỹ':
			result.WriteRune('Y')
		case 'đ':
			result.WriteRune('d')
		case 'Đ':
			result.WriteRune('D')
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}

func NewProductController(service service.ProductService, s3Service service.S3Service) *ProductController {
	return &ProductController{
		service:   service,
		s3Service: &s3Service,
	}
}

// generateVariantID - Tạo ID duy nhất cho variant dựa trên productID, size và color
func (ctrl *ProductController) generateVariantID(productID, size, color string) string {
	// Bỏ dấu và chuẩn hóa size và color
	normalizedSize := strings.ToUpper(strings.TrimSpace(removeDiacritics(size)))
	normalizedColor := strings.ToUpper(strings.TrimSpace(removeDiacritics(color)))

	// Thay thế khoảng trắng bằng dấu gạch ngang
	normalizedColor = strings.ReplaceAll(normalizedColor, " ", "-")
	normalizedSize = strings.ReplaceAll(normalizedSize, " ", "-")

	// Format: PRODUCT-ID-SIZE-COLOR
	return fmt.Sprintf("%s-%s-%s", productID, normalizedSize, normalizedColor)
}

// getCategoryCode trả về mã category chuẩn
func getCategoryCode(categoryName string) string {
	// Mapping tên category sang code
	categoryMap := map[string]string{
		"Áo thun":        "SHIRT",
		"Áo polo":        "POLO",
		"Quần jeans":     "JEANS",
		"Quần short":     "SHORT",
		"Giày thể thao":  "SHOES",
		"Giày da":        "SHOES-L",
		"Túi xách":       "BAG",
		"Balo":           "BACKPACK",
		"Điện thoại":     "PHONE",
		"Laptop":         "LAPTOP",
		"Phụ kiện":       "ACCESS",
		"Đồng hồ":        "WATCH",
		"Mỹ phẩm":        "COSMETIC",
		"Thời trang nam": "F-MEN",
		"Thời trang nữ":  "F-WOM",
		"Đồ gia dụng":    "HOME",
	}

	// Tìm code từ map
	if code, ok := categoryMap[categoryName]; ok {
		return code
	}

	// Nếu không có trong map, chuẩn hóa tên category
	normalized := removeDiacritics(categoryName)
	return strings.ToUpper(strings.ReplaceAll(normalized, " ", "-"))
}

// generateProductID tạo ID duy nhất cho product dựa trên category
func (ctrl *ProductController) generateProductID(ctx context.Context, category string) string {
	// Lấy category code
	categoryCode := getCategoryCode(category)

	// Tạo 5 ký tự ngẫu nhiên
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 5)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	randomPart := string(b)

	return fmt.Sprintf("SKU-%s-%s", categoryCode, randomPart)
}

// randomString - Tạo chuỗi ngẫu nhiên
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func (ctrl *ProductController) AddProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()
		// CheckSellerRole(c)
		if c.IsAborted() {
			return
		}

		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found"})
			return
		}
		var req models.CreateProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
			return
		}

		// Handle empty ImagePath - DynamoDB doesn't allow empty string sets
		imagePath := req.ImagePath
		if len(imagePath) == 0 {
			imagePath = nil
		}

		// Tạo variants từ request (chưa có ID, service sẽ generate)
		variants := make([]models.ProductVariant, len(req.Variants))
		for i, v := range req.Variants {
			variants[i] = models.ProductVariant{
				Size:      v.Size,
				Color:     v.Color,
				Material:  v.Material,
				CostPrice: v.CostPrice,
				Price:     v.Price,
				Quantity:  v.Quantity,
			}
		}

		product := models.Product{
			Name:        req.Name,
			Category:    req.Category,
			Description: req.Description,
			ImagePath:   imagePath,
			UserID:      userID,
			Status:      req.Status,
			Variants:    variants,
		}

		if err := ctrl.service.AddProduct(ctx, product); err != nil {
			logger.Error("Error adding product", zap.Error(err))
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

		var req models.UpdateProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Error binding JSON for EditProduct", zap.Error(err))
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
		if req.Status != nil {
			update["status"] = *req.Status
		}

		// Cập nhật variants
		if len(req.Variants) > 0 {
			// Lấy product hiện tại để merge variants
			currentProduct, err := ctrl.service.GetProductByID(ctx, id)
			if err != nil {
				logger.Error("Error fetching current product for variant update", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
				return
			}

			// Tạo map để tra cứu variants hiện tại
			variantMap := make(map[string]models.ProductVariant)
			for _, v := range currentProduct.Variants {
				variantMap[v.ID] = v
			}

			// Update hoặc thêm mới variants
			for _, v := range req.Variants {
				if v.ID != "" {
					// Update variant hiện có
					if existing, ok := variantMap[v.ID]; ok {
						if v.Size != nil {
							existing.Size = *v.Size
						}
						if v.Color != nil {
							existing.Color = *v.Color
						}
						if v.Material != nil {
							existing.Material = *v.Material
						}
						if v.Price != nil {
							existing.Price = *v.Price
						}
						if v.Quantity != nil {
							existing.Quantity = *v.Quantity
						}
						variantMap[v.ID] = existing
					}
				} else {
					// Thêm variant mới - cần size và color để generate ID
					if v.Size == nil || v.Color == nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": "Size and color are required for new variants"})
						return
					}

					newVariant := models.ProductVariant{
						ID:        ctrl.generateVariantID(id, *v.Size, *v.Color),
						CreatedAt: time.Now(),
					}
					if v.Size != nil {
						newVariant.Size = *v.Size
					}
					if v.Color != nil {
						newVariant.Color = *v.Color
					}
					if v.Material != nil {
						newVariant.Material = *v.Material
					}
					if v.CostPrice != nil {
						newVariant.CostPrice = *v.CostPrice
					}
					if v.Price != nil {
						newVariant.Price = *v.Price
					}
					if v.Quantity != nil {
						newVariant.Quantity = *v.Quantity
					}
					variantMap[newVariant.ID] = newVariant
				}
			}

			// Chuyển map thành array
			variants := make([]models.ProductVariant, 0, len(variantMap))
			for _, v := range variantMap {
				variants = append(variants, v)
			}
			update["variants"] = variants
		}

		if len(update) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			logger.Error("Failed to update product: no fields provided")
			return
		}

		categoryName := ""
		if req.Category != nil {
			categoryName = *req.Category
		}

		if err := ctrl.service.EditProduct(ctx, id, update, categoryName); err != nil {
			logger.Error("Error updating product", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
	}
}

func (ctrl *ProductController) DeleteProductVariant() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		productID := c.Param("id")
		variantID := c.Param("variant_id")

		if productID == "" || variantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID and Variant ID are required"})
			return
		}

		// Lấy product hiện tại
		product, err := ctrl.service.GetProductByID(ctx, productID)
		if err != nil {
			logger.Error("Error fetching product", zap.Error(err))
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		// Lọc bỏ variant cần xóa
		newVariants := make([]models.ProductVariant, 0)
		found := false
		for _, v := range product.Variants {
			if v.ID != variantID {
				newVariants = append(newVariants, v)
			} else {
				found = true
			}
		}

		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
			return
		}

		if len(newVariants) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete last variant. Product must have at least one variant"})
			return
		}

		// Update product với variants mới
		update := map[string]interface{}{
			"variants": newVariants,
		}

		if err := ctrl.service.EditProduct(ctx, productID, update, ""); err != nil {
			logger.Error("Error deleting variant", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete variant"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Variant deleted successfully"})
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

		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found"})
			logger.Error("User ID not found in DeleteProduct")
			return
		}
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID not found"})
			logger.Error("Product ID not found in DeleteProduct")
			return
		}

		if err := ctrl.service.DeleteProduct(ctx, id, userID); err != nil {
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
	VariantID string `json:"variant_id"` // ID của variant cần update
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

		// Call UpdateProductStock với product ID, variant ID và quantity
		err := ctrl.service.UpdateProductStock(ctx, item.ProductID, item.VariantID, quantity)
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

func (ctrl *ProductController) GetProductByUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		userID := c.GetHeader("X-User-ID")
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
		products, total, pages, hasNext, hasPrev, err := ctrl.service.GetProductByUserID(ctx, userID, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := gin.H{
			"data":     products,
			"total":    total,
			"page":     page,
			"pages":    pages,
			"has_next": hasNext,
			"has_prev": hasPrev,
		}

		c.JSON(http.StatusOK, response)
	}
}

func (ctrl *ProductController) GetProductByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID not found"})
			return
		}

		product, err := ctrl.service.GetProductByID(ctx, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if product == nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
			return
		}

		c.JSON(http.StatusOK, product)
	}
}

func (ctrl *ProductController) GetProductByCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		category := c.Param("category")
		if category == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category not found"})
			return
		}

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

		products, total, pages, hasNext, hasPrev, err := ctrl.service.GetProductByCategory(ctx, category, page, limit)
		if err != nil {
			logger.Error("failed to get product by category", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := gin.H{
			"data":     products,
			"total":    total,
			"page":     page,
			"pages":    pages,
			"has_next": hasNext,
			"has_prev": hasPrev,
		}

		c.JSON(http.StatusOK, response)
	}
}

func (ctrl *ProductController) AddProductCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		userType := c.GetHeader("X-User-Type")
		if userType != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		var req struct {
			Name string `json:"name" binding:"required,min=2,max=100"`
			Code string `json:"code" binding:"omitempty,alphanum,max=20"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
			return
		}

		category := models.Category{
			Name:      req.Name,
			Code:      req.Code,
			CreatedAt: time.Now(),
		}
		err := ctrl.service.AddProductCategory(ctx, category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Category added successfully"})
	}
}

func (ctrl *ProductController) GetProductCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		categories, err := ctrl.service.GetProductCategory(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": categories})
	}
}

func (ctrl *ProductController) DeleteProductCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		userType := c.GetHeader("X-User-Type")
		if userType != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		categoryID := c.Param("id")
		if categoryID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID not found"})
			return
		}

		err := ctrl.service.DeleteProductCategory(ctx, categoryID)
		if err != nil {
			logger.Err("error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
	}
}

func (ctrl *ProductController) GetProductStatistics() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		userType := c.GetHeader("X-User-Type")
		if userType != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		monthStr := c.Query("month")
		yearStr := c.Query("year")
		var month, year int
		var err error

		if monthStr != "" {
			month, err = strconv.Atoi(monthStr)
			if err != nil || month < 1 || month > 12 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month; must be between 1 and 12"})
				return
			}
		}

		if yearStr != "" {
			year, err = strconv.Atoi(yearStr)
			if err != nil || year < 1970 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
				return
			}
		}

		if month > 0 && year == 0 {
			year = time.Now().Year()
		}

		stats, err := ctrl.service.GetProductStatistics(ctx, month, year)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": stats})
	}
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
