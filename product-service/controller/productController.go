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
	"strconv"
	"sync"
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


func saveImageToFileSystem(c *gin.Context, file *multipart.FileHeader) (string, error) {
	saveDir := "./product-service/uploads/images/"
	err := os.MkdirAll(saveDir, os.ModePerm) // Đảm bảo thư mục tồn tại
	if err != nil {
		return "", fmt.Errorf("Failed to create directory: %v", err)
	}

	// Tạo tên file duy nhất
	imageFileName := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)
	imagePath := saveDir + imageFileName

	// Lưu file vào thư mục
	if err := c.SaveUploadedFile(file, imagePath); err != nil {
		return "", fmt.Errorf("Failed to save image: %v", err)
	}

	return imagePath, nil
}

                                   

func AddProduct() gin.HandlerFunc {
    return func(c *gin.Context) {

        log.Printf("Content-Type: %s", c.GetHeader("Content-Type"))
        log.Printf("All form values: %v", c.Request.Form)

        CheckUserRole(c)
        if c.IsAborted(){
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


func EditProduct() gin.HandlerFunc{
	return func (c *gin.Context)  {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		productID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product"})
			return
		}

		CheckUserRole(c)

		name := c.PostForm("name")
		description := c.PostForm("description")
		priceStr := c.PostForm("price")
		quantityStr := c.PostForm("quantity")

		update := bson.M{"updated_at": time.Now()}
		if name != ""{
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
		if err == nil{
			imagePath, err := saveImageToFileSystem(c,file)
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			update["image_path"] = imagePath
		}

		result, err := productCollection.UpdateOne(
			ctx,
			bson.M{"_id":productID},
			bson.M{"$set": update},
		)
		
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}

		if result.MatchedCount == 0{
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
            "_id":     objID,
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
            "id": productID,
        })
    }
}

func GetProductByName(name string) ([]models.Product, error){
	var products []models.Product
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

	// filter := bson.D{{"name", bson.D{{"$regex", name}, {"$options", "i"}}}}
	filter := bson.M{"name": bson.M{"$regex": name, "$options": "i"}}
	cursor, err := productCollection.Find(ctx, filter)
	if  err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx){
		var product models.Product
		if err := cursor.Decode(&product); err != nil{
			return nil, err
		}
		products = append(products, product)
	}

	if err := cursor.Err(); err != nil{
		return nil, err
	}
	
	defer cancel()
	return products, nil

}

func GetProdctByNameHander() gin.HandlerFunc{
	return func (c *gin.Context)  {
		name := c.Query("name")
		// CheckUserRole(c)
		if name == ""{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name query parameter is required"})
			return
		}

		products, err := GetProductByName(name)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(products) == 0{
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

func GetAllProducts() gin.HandlerFunc{
    return func (c *gin.Context)  {
        log.Printf("Starting GetAllProducts handler")
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()

        page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
        if err != nil || page < 1{
            page = 1
        }
        limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
        if err != nil || limit < 1{
            limit = 10
        }

        cacheKey := fmt.Sprintf("products_%d_%d", page, limit)

        type cacheResult struct {
            found bool
            products []models.Product
            total int64
            err error
        }

        cacheChannel := make(chan cacheResult, 1)

        go func(){
            var result cacheResult
            cacheData , err := database.RedisClient.Get(ctx, cacheKey).Result()
            if err == nil{
                var cachedResponse struct {
                    Products []models.Product `json:"products"`
                    Total int64 `json:"total"`
                }

                if err := json.Unmarshal([]byte(cacheData), &cachedResponse); err == nil{
                    result.found = true
                    result.products = cachedResponse.Products
                    result.total = cachedResponse.Total
                }
            }
            cacheChannel <- result
        }()

        result := <- cacheChannel

        if result.found{
            total := result.total
            product:= result.products

            pages := int(math.Ceil(float64(total) / float64(limit)))

            c.JSON(http.StatusOK, gin.H{
                "data": product,
                "total": total,
                "page": page,
                "pages": pages,
                "has_next": page < pages,
                "has_prev": page > 1,
                "source": "cache",
            })
            return
        }

        var wg sync.WaitGroup
        wg.Add(2)

        type countResult struct {
            total int64
            err error
        }

        type productResult struct {
            products []models.Product
            err error
        }

        countChan := make(chan countResult, 1)
        productsChan := make(chan productResult, 1)

        go func(){
            defer wg.Done()

            total, err := productCollection.CountDocuments(ctx, bson.M{})
            countChan <- countResult{total, err}
        }()

        go func(){
            defer wg.Done()

            skip := (page - 1) * limit

            findOptions := options.Find().
                SetSkip(int64(skip)).
                SetLimit(int64(limit)).
                SetSort(bson.D{{"created_at", -1}})

            cursor, err := productCollection.Find(ctx, bson.M{}, findOptions)
            if err != nil{
                productsChan <- productResult{err: err}
                return
            }

            defer cursor.Close(ctx)

            var products []models.Product
            if err := cursor.All(ctx, &products); err != nil{
                productsChan <- productResult{err:err}
                return
            }

            productsChan <- productResult{products: products}
        }()

        go func(){
            wg.Wait()
            close(countChan)
            close(productsChan)
        }()

        countRes := <- countChan
        if countRes.err != nil{
            log.Printf("Error counting products: %v", countRes.err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count products"})
            return
        }

        productRes := <- productsChan
        if productRes.err != nil{
            log.Printf("Error fetching products: %v", productRes.err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
            return
        }

        total := countRes.total
        products := productRes.products


        if total == 0 {
            c.JSON(http.StatusOK, gin.H{
                "data":     []models.Product{},
                "total":    0,
                "page":     page,
                "pages":    0,
                "has_next": false,
                "has_prev": false,
                "source": "db",
            })
            return
        }

        pages := int(math.Ceil(float64(total) / float64(limit)))

        response := gin.H{
            "data":     products,
            "total":    total,
            "page":     page,
            "pages":    pages,
            "has_next": page < pages,
            "has_prev": page > 1,
            "source": "database",
        }

        cacheData := struct {
            Products []models.Product `json:"products"`
            Total int64 `json:"total"`
        }{
            Products: products,
            Total: total,
        }

        cacheJSON, err := json.Marshal(cacheData)
        if err != nil{
            database.RedisClient.Set(ctx, cacheKey, cacheJSON, 10*time.Minute)
        }

        c.JSON(http.StatusOK, response)
    }

}


func CheckStock(productID string) (int, error){

    filter := bson.M{"_id": productID}

    var product models.Product

    err := productCollection.FindOne(context.Background(), filter).Decode(&product)
    if err != nil{
        return 0, fmt.Errorf("Product not found:" )
    }
    return *product.Quantity, nil
}
