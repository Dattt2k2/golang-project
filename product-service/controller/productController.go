package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// "sync"
	"time"

	"github.com/Dattt2k2/golang-project/product-service/database"
	"github.com/Dattt2k2/golang-project/product-service/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var productCollection *mongo.Collection = database.OpenCollection(database.Client, "product")
var validate = validator.New()

func CheckUserRole(c *gin.Context) {
	userRole := c.GetHeader("user_type")
	if userRole != "SELLER" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You don't have permission"})
		c.Abort()
		return
	}
	c.Next()
}

// func saveImageToFileSystem(c *gin.Context, file *multipart.FileHeader) (string, error) {
// 	saveDir := "./product-service/uploads/images/"
// 	err := os.MkdirAll(saveDir, os.ModePerm) // Đảm bảo thư mục tồn tại
// 	if err != nil {
// 		return "", fmt.Errorf("Failed to create directory: %v", err)
// 	}

// 	// Tạo tên file duy nhất
// 	imageFileName := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)
// 	imagePath := saveDir + imageFileName

// 	// Lưu file vào thư mục
// 	if err := c.SaveUploadedFile(file, imagePath); err != nil {
// 		return "", fmt.Errorf("Failed to save image: %v", err)
// 	}

// 	return imagePath, nil
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

func AddProduct() gin.HandlerFunc {
	return func(c *gin.Context) {

		log.Printf("Content-Type: %s", c.GetHeader("Content-Type"))
		log.Printf("All form values: %v", c.Request.Form)

		CheckUserRole(c)
		if c.IsAborted() {
			return
		}

		// Parse multipart form
		if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
			log.Printf("Error parsing multipart form: %v", err)
		}

		// Get form data
		formData := c.Request.MultipartForm
		log.Printf("Form data: %v", formData)

		var product models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Get user ID from header and check role
		userID := c.GetHeader("user_id")

		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Get form values
		name := c.PostForm("name")
		description := c.PostForm("description")
		quantityStr := c.PostForm("quantity")
		priceStr := c.PostForm("price")
		category := c.PostForm("category")

		log.Printf("Received quantity: %s", quantityStr)

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

		// Handle image file
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
			return
		}

		imagePath, err := saveImageToFileSystem(c, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Create product object
		product = models.Product{
			ID:          primitive.NewObjectID(),
			Name:        &name,
			Category:    &category,
			Description: &description,
			Price:       price,
			Quantity:    &quantity,
			ImagePath:   imagePath,
			Created_at:  time.Now(),
			Updated_at:  time.Now(),
			UserID:      userObjectID, // Set the user ID from header
		}

		// Validate product
		if err := validate.Struct(product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Product: %v", product)

		// Insert product into MongoDB
		_, err = productCollection.InsertOne(ctx, product)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product added successfully"})
	}
}
func EditProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		productID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product"})
			return
		}

		CheckUserRole(c)

		name := c.PostForm("name")
		description := c.PostForm("description")
		priceStr := c.PostForm("price")
		quantityStr := c.PostForm("quantity")

		update := bson.M{"updated_at": time.Now()}
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

		result, err := productCollection.UpdateOne(
			ctx,
			bson.M{"_id": productID},
			bson.M{"$set": update},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
	}
}

func DeleteProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		log.Printf("Starting DeleteProduct handler")

		productID := c.Param("id")
		log.Printf("Product ID from URL: %s", productID)

		objID, err := primitive.ObjectIDFromHex(productID)
		if err != nil {
			log.Printf("Error converting product ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		var product bson.M
		err = productCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
		log.Printf("Product before delete: %+v", product)

		userID := c.GetHeader("user_id")
		if userID == "" {
			log.Printf("User ID not found in header")
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found"})
			return
		}

		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			log.Printf("Error converting user ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		filter := bson.M{
			"_id":    objID,
			"userid": userObjectID,
		}
		log.Printf("Delete filter: %v", filter)

		result, err := productCollection.DeleteOne(ctx, filter)
		if err != nil {
			log.Printf("Error deleting product: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
			return
		}

		if result.DeletedCount == 0 {
			log.Printf("Product not found or unauthorized. Filter: %v", filter)
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found or you don't have permission to delete it"})
			return
		}

		log.Printf("Successfully deleted product. ProductID: %s, UserID: %s", productID, userID)
		c.JSON(http.StatusOK, gin.H{
			"message": "Product deleted successfully",
			"id":      productID,
		})
	}
}

func GetProductByName(name string) ([]models.Product, error) {
	var products []models.Product
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// filter := bson.D{{"name", bson.D{{"$regex", name}, {"$options", "i"}}}}
	filter := bson.M{"name": bson.M{"$regex": name, "$options": "i"}}
	cursor, err := productCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// defer cancel()
	return products, nil

}

func GetProdctByNameHander() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")
		// CheckUserRole(c)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name query parameter is required"})
			return
		}

		products, err := GetProductByName(name)
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

// func GetAllProducts() gin.HandlerFunc {
//     return func(c *gin.Context) {
//         log.Printf("Starting GetAllProducts handler")

//         var products []models.Product
//         var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
//         defer cancel()

//         // CheckUserRole(c)
//         if c.IsAborted(){
//             return
//         }

//         page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
//         if err != nil || page < 1 {
//             log.Printf("Invalid page parameter, using default: %v", err)
//             page = 1
//         }
//         limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
//         if err != nil || limit < 1 {
//             log.Printf("Invalid limit parameter, using default: %v", err)
//             limit = 10
//         }

//         log.Printf("Pagination: page=%d, limit=%d", page, limit)

//         skip := (page - 1) * limit

//         // Count total products
//         total, err := productCollection.CountDocuments(ctx, bson.M{})
//         if err != nil {
//             log.Printf("Error counting products: %v", err)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count products"})
//             return
//         }
//         log.Printf("Total products count: %d", total)

//         if total == 0 {
//             c.JSON(http.StatusOK, gin.H{
//                 "data":     []models.Product{},
//                 "total":    0,
//                 "page":     page,
//                 "pages":    0,
//                 "has_next": false,
//                 "has_prev": false,
//             })
//             return
//         }

//         // Create find options
//         findOptions := options.Find().
//             SetSkip(int64(skip)).
//             SetLimit(int64(limit)).
//             SetSort(bson.D{{"created_at", -1}})

//         // Find products
//         cursor, err := productCollection.Find(ctx, bson.M{}, findOptions)
//         if err != nil {
//             log.Printf("Error fetching products: %v", err)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
//             return
//         }
//         defer cursor.Close(ctx)

//         // Decode products
//         if err := cursor.All(ctx, &products); err != nil {
//             log.Printf("Error decoding products: %v", err)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse products"})
//             return
//         }

//         log.Printf("Found %d products for current page", len(products))

//         // Calculate pagination info
//         pages := int(math.Ceil(float64(total) / float64(limit)))

//         // Add debug info to verify product data
//         for i, product := range products {
//             log.Printf("Product %d: ID=%v, Name=%v", i, product.ID, *product.Name)
//         }

//         response := gin.H{
//             "data":     products,
//             "total":    total,
//             "page":     page,
//             "pages":    pages,
//             "has_next": page < pages,
//             "has_prev": page > 1,
//         }

//         log.Printf("Sending response with %d products", len(products))
//         c.JSON(http.StatusOK, response)
//     }
// }

func GetAllProducts() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Starting GetAllProducts handler")

		// Check if Redis client is initialized
		if database.RedisClient == nil {
			log.Printf("WARNING: Redis client is nil, initializing...")
			database.InitRedis() // Make sure this function exists

			// If still nil after init, proceed without Redis
			if database.RedisClient == nil {
				log.Printf("ERROR: Failed to initialize Redis client, proceeding without caching")
				// Continue with MongoDB only approach
			}
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		if c.IsAborted() {
			return
		}

		// Parse pagination parameters
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			log.Printf("Invalid page parameter, using default: %v", err)
			page = 1
		}
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil || limit < 1 {
			log.Printf("Invalid limit parameter, using default: %v", err)
			limit = 10
		}

		log.Printf("Pagination: page=%d, limit=%d", page, limit)

		// Create cache key
		cacheKey := fmt.Sprintf("products_page_%d_limit_%d", page, limit)
		log.Printf("Cache key: %s", cacheKey)

		// Try to get data from Redis cache
		cachedData, err := database.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			log.Printf("Cache hit for key: %s", cacheKey)

			// Unmarshal cached data
			var cachedResponse gin.H
			if err := json.Unmarshal([]byte(cachedData), &cachedResponse); err == nil {
				log.Printf("Successfully unmarshaled cached data")

				// Add cache source information
				cachedResponse["cache"] = true

				c.JSON(http.StatusOK, cachedResponse)
				return
			} else {
				log.Printf("Error unmarshaling cached data: %v", err)
			}
		} else {
			log.Printf("Cache miss for key: %s, error: %v", cacheKey, err)
		}

		// If cache miss or error, fetch from MongoDB
		skip := (page - 1) * limit

		// Count total products
		total, err := productCollection.CountDocuments(ctx, bson.M{})
		if err != nil {
			log.Printf("Error counting products: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count products"})
			return
		}
		log.Printf("Total products count: %d", total)

		if total == 0 {
			emptyResponse := gin.H{
				"data":     []models.Product{},
				"total":    0,
				"page":     page,
				"pages":    0,
				"has_next": false,
				"has_prev": false,
				"cache":    false,
			}

			// Cache empty result
			cacheData, _ := json.Marshal(emptyResponse)
			database.RedisClient.Set(ctx, cacheKey, cacheData, 10*time.Minute)

			c.JSON(http.StatusOK, emptyResponse)
			return
		}

		// Create find options
		findOptions := options.Find().
			SetSkip(int64(skip)).
			SetLimit(int64(limit)).
			SetSort(bson.D{{"created_at", -1}})

		// Find products
		var products []models.Product
		cursor, err := productCollection.Find(ctx, bson.M{}, findOptions)
		if err != nil {
			log.Printf("Error fetching products: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}
		defer cursor.Close(ctx)

		// Decode products
		if err := cursor.All(ctx, &products); err != nil {
			log.Printf("Error decoding products: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse products"})
			return
		}

		log.Printf("Found %d products for current page", len(products))

		// Calculate pagination info
		pages := int(math.Ceil(float64(total) / float64(limit)))

		// Add debug info to verify product data
		for i, product := range products {
			log.Printf("Product %d: ID=%v, Name=%v", i, product.ID, *product.Name)
		}

		response := gin.H{
			"data":     products,
			"total":    total,
			"page":     page,
			"pages":    pages,
			"has_next": page < pages,
			"has_prev": page > 1,
			"cache":    false,
		}

		// Cache the response for future requests
		cacheData, err := json.Marshal(response)
		if err == nil {
			// Set cache with expiration time
			err = database.RedisClient.Set(ctx, cacheKey, cacheData, 10*time.Minute).Err()
			if err != nil {
				log.Printf("Error caching products: %v", err)
			} else {
				log.Printf("Successfully cached products with key: %s", cacheKey)
			}
		} else {
			log.Printf("Error marshaling products for cache: %v", err)
		}

		log.Printf("Sending response with %d products", len(products))
		c.JSON(http.StatusOK, response)
	}
}

func CheckStock(productID string) (int, error) {

	filter := bson.M{"_id": productID}

	var product models.Product

	err := productCollection.FindOne(context.Background(), filter).Decode(&product)
	if err != nil {
		return 0, fmt.Errorf("Product not found:")
	}
	return *product.Quantity, nil
}

func GetProductImage() gin.HandlerFunc {
    return func(c *gin.Context) {
        filename := c.Param("filename")
        log.Printf("Requested image: %s", filename)
        
        // Get current working directory
        wd, err := os.Getwd()
        if err != nil {
            log.Printf("Error getting working directory: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
            return
        }
        
        // Try many possible paths
        possiblePaths := []string{
            filepath.Join(wd, "product-service", "uploads", "images", filename),
            filepath.Join(wd, "uploads", "images", filename),
            filepath.Join("product-service", "uploads", "images", filename),
            filepath.Join("uploads", "images", filename),
            filename, // Try just the filename
            filepath.Join("/tmp", filename),
            filepath.Join("/app", "uploads", "images", filename),
            filepath.Join("/app", "product-service", "uploads", "images", filename),
        }
        
        // Check each path
        var foundPath string
        for _, path := range possiblePaths {
            log.Printf("Checking path: %s", path)
            if _, err := os.Stat(path); err == nil {
                foundPath = path
                log.Printf("Image found at: %s", foundPath)
                break
            } else {
                log.Printf("Image not found at: %s", path)
            }
        }
        
        if foundPath == "" {
            // Do a recursive search for the file as a last resort
            log.Printf("Starting recursive search for file: %s", filename)
            foundPath = searchFileRecursively(wd, filename)
            
            if foundPath == "" {
                log.Printf("Image not found in any location: %s", filename)
                c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
                return
            }
        }
        
        // Determine content type based on file extension
        ext := strings.ToLower(filepath.Ext(filename))
        var contentType string
        switch ext {
        case ".jpg", ".jpeg":
            contentType = "image/jpeg"
        case ".png":
            contentType = "image/png"
        case ".gif":
            contentType = "image/gif"
        default:
            contentType = "application/octet-stream"
        }
        
        log.Printf("Serving image: %s with content type: %s", foundPath, contentType)
        c.Header("Content-Type", contentType)
        c.File(foundPath)
    }
}

// Helper function to recursively search for a file
func searchFileRecursively(rootDir string, filename string) string {
    log.Printf("Searching in directory: %s", rootDir)
    
    files, err := os.ReadDir(rootDir)
    if err != nil {
        log.Printf("Error reading directory %s: %v", rootDir, err)
        return ""
    }
    
    for _, file := range files {
        if file.IsDir() {
            // Skip certain directories to avoid endless recursion
            if file.Name() == "node_modules" || file.Name() == ".git" {
                continue
            }
            
            // Recursively search subdirectory
            path := searchFileRecursively(filepath.Join(rootDir, file.Name()), filename)
            if path != "" {
                return path
            }
        } else if file.Name() == filename {
            // Found the file
            path := filepath.Join(rootDir, file.Name())
            log.Printf("File found at: %s", path)
            return path
        }
    }
    
    return ""
}

// Add this to your productController.go in the product service
func FindImageLocations() gin.HandlerFunc {
    return func(c *gin.Context) {
        result := make(map[string]interface{})
        
        // Get the current working directory
        wd, err := os.Getwd()
        if err != nil {
            result["error"] = fmt.Sprintf("Error getting working directory: %v", err)
            c.JSON(http.StatusInternalServerError, result)
            return
        }
        
        result["working_directory"] = wd
        
        // List of directories to search
        dirsToSearch := []string{
            wd,
            filepath.Join(wd, "product-service"),
            filepath.Join(wd, "uploads"),
            filepath.Join(wd, "product-service", "uploads"),
            filepath.Join(wd, "product-service", "uploads", "images"),
            filepath.Join(wd, "uploads", "images"),
            "/tmp",
            "/app",
            "/app/product-service",
            "/app/uploads",
            "/app/product-service/uploads",
            "/app/product-service/uploads/images",
            "/app/uploads/images",
        }
        
        // Search for image files in these directories
        foundImages := make(map[string][]string)
        
        for _, dir := range dirsToSearch {
            // Check if directory exists
            info, err := os.Stat(dir)
            if os.IsNotExist(err) || !info.IsDir() {
                continue
            }
            
            // Read files in the directory
            files, err := os.ReadDir(dir)
            if err != nil {
                continue
            }
            
            // Filter for image files
            imageFiles := []string{}
            for _, file := range files {
                if !file.IsDir() {
                    // Check if it's an image file by extension
                    ext := strings.ToLower(filepath.Ext(file.Name()))
                    if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
                        imageFiles = append(imageFiles, file.Name())
                    }
                }
            }
            
            if len(imageFiles) > 0 {
                foundImages[dir] = imageFiles
            }
        }
        
        result["found_images"] = foundImages
        
        // Also try to locate the specific file
        searchFilename := "1742719764-FinWell.png" // The problematic image
        foundPaths := []string{}
        
        for _, dir := range dirsToSearch {
            fullPath := filepath.Join(dir, searchFilename)
            if _, err := os.Stat(fullPath); err == nil {
                foundPaths = append(foundPaths, fullPath)
            }
        }
        
        result["specific_image_found_at"] = foundPaths
        
        c.JSON(http.StatusOK, result)
    }
}

// Add this function to verify if images exist
func VerifyImageExists() gin.HandlerFunc {
    return func(c *gin.Context) {
        wd, err := os.Getwd()
        if err != nil {
            log.Printf("Error getting working directory: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
            return
        }
        
        // Use absolute paths
        uploadDirs := []string{
            filepath.Join(wd, "uploads", "images"),
            filepath.Join(wd, "product-service", "uploads", "images"),
            filepath.Join(wd, "..", "uploads", "images"),
        }
        
        result := make(map[string]interface{})
        
        for _, dir := range uploadDirs {
            dirInfo := make(map[string]interface{})
            
            // Check if directory exists
            if _, err := os.Stat(dir); os.IsNotExist(err) {
                dirInfo["exists"] = false
                dirInfo["error"] = "Directory does not exist"
            } else {
                dirInfo["exists"] = true
                
                // Try to read files
                files, err := os.ReadDir(dir)
                if err != nil {
                    dirInfo["readable"] = false
                    dirInfo["error"] = fmt.Sprintf("Cannot read directory: %v", err)
                } else {
                    dirInfo["readable"] = true
                    
                    fileList := make([]string, 0)
                    for _, file := range files {
                        fileList = append(fileList, file.Name())
                    }
                    
                    dirInfo["files"] = fileList
                    dirInfo["count"] = len(fileList)
                }
            }
            
            // Add to result
            result[dir] = dirInfo
        }
        
        // Add working directory for reference
        result["working_directory"] = wd
        
        c.JSON(http.StatusOK, result)
    }
}