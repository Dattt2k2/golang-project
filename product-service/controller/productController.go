package controllers

import (
	"context"
	"fmt"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
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
	userRole := c.GetHeader("role")
	if userRole != "SELLER" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You don't have permission"})
		c.Abort()
	}
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

func AddProduct(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {

		var product models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID := c.GetHeader("uid")
		CheckUserRole(c)

		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		name := c.PostForm("name")
		description := c.PostForm("description")
		quantityStr := c.PostForm("quantity")
		priceStr := c.PostForm("price")

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

		// Xử lý file ảnh
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
			return
		}

		imagePath, err := saveImageToFileSystem(c,file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		product.ID = primitive.NewObjectID()
		product.Name = &name
		product.ImagePath = imagePath
		product.Description = &description
		product.Price = price
		product.Quantity = &quantity
		product.Created_at = time.Now()
		product.Updated_at = time.Now()
		product.UserID = userObjectID

		// Validate dữ liệu
		if err := validate.Struct(product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Chèn dữ liệu vào MongoDB
		_, err = db.Collection("product").InsertOne(ctx, product)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product added successfully"})
	}
}


// func parseFloat(s string) float64{
// 	f, _ := strconv.ParseFloat(s, 64)
// 	return f
// }


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


// func handleImageUpload(ctx, context.Context, productID primitive.ObjecID, file *multipart.FileHeader) (primitive.ObjectID, error){
// 	var product models.Product

// 	err := productCollection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product)
// 	if err != nil{
// 		return primitive.NilObjectID, fmt.Errorf("Failed to fetch product")
// 	}

// 	if product.Image_id != primitive.NilObjectID{
// 		err := database.GetBucket().Delete(product.)
// 	}
// }



func DeleteProduct() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		productID := c.Param("id")

		CheckUserRole(c)

		objID, _ := primitive.ObjectIDFromHex(productID)

		result, err := productCollection.DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete product"})
			return
		}
		
		if result.DeletedCount == 0{
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "Delete product complete"})
		defer cancel()
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
		CheckUserRole(c)
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

func GetAllProducts(db *mongo.Database) gin.HandlerFunc{
	return func(c *gin.Context) {
		var products []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		CheckUserRole(c)
		// Lấy tham số page và limit từ query
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			page = 1
		}
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil || limit < 1 {
			limit = 10
		}

		// Tính toán skip và limit
		skip := (page - 1) * limit

		// Tổng số sản phẩm
		total, err := db.Collection("products").CountDocuments(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count products"})
			return
		}

		// Lấy sản phẩm từ MongoDB
		cursor, err := db.Collection("products").Find(ctx, bson.M{}, options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}
		defer cursor.Close(ctx)

		// Decode dữ liệu
		if err := cursor.All(ctx, &products); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse products"})
			return
		}

		// Tổng số trang
		pages := int(math.Ceil(float64(total) / float64(limit)))

		// Trả dữ liệu và metadata
		c.JSON(http.StatusOK, gin.H{
			"data":  products,
			"total": total,
			"page":  page,
			"pages": pages,
		})
	}
}

