package controllers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	database "github.com/Dattt2k2/golang-project/database/databaseConnection.gp"
	"github.com/Dattt2k2/golang-project/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)


var productCollection *mongo.Collection = database.OpenCollection(database.Client, "product")
var validate = validator.New()

// func AddProduct() gin.HandlerFunc{
// 	return func(c *gin.Context){
// 		var product models.Product
// 		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
// 		if err := c.BindJSON(&product); err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}
// 		validationErr := validate.Struct(product)
// 		if validationErr != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
// 			return
// 		}
		
// 		product.Created_at = time.Now()
// 		product.Updated_at = time.Now()
// 		product.ID = primitive.NewObjectID()

// 		result, err := productCollection.InsertOne(ctx, product)
// 		if err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}
// 		c.JSON(http.StatusOK, gin.H{"id": result.InsertedID})
// 		defer cancel()
// 	}
// }

// var db *mongo.Database 

func AddProduct() gin.HandlerFunc{
	return func(c *gin.Context){
		var product models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		
		if err:= c.ShouldBindJSON(&product); err!= nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(product)
		if validationErr != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		file, header, err := c.Request.FormFile("image")
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload image"})
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read all file"})
			return
		}
		
		bucket, err := gridfs.NewBucket(
			productCollection.Database(),
			options.GridFSBucket().SetName("images"),
		)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create Gridfs bucket"})
			return
		}

		uploadStream, err := bucket.OpenUploadStream(header.Filename)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload stream"})
			return
		}
		defer uploadStream.Close()

		_, err = uploadStream.Write(data)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file to GridFs"})
			return
		}

		fileID, ok := uploadStream.FileID.(primitive.ObjectID)
		if !ok{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convet FildId to ObjectId"})
			return
		}

		//get userID
		userID := c.GetString("uid")
		if userID == ""{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get userID"})
			return
		}

		product.UserID, _ = primitive.ObjectIDFromHex(userID)
		product.Image = fileID.Hex()
		product.Created_at = time.Now()
		product.Updated_at = time.Now()
		product.ID = primitive.NewObjectID()

		result, err := productCollection.InsertOne(ctx, product)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}


		c.JSON(http.StatusOK, gin.H{"id": result.InsertedID})
		defer cancel()
	}
}

// func EditProduct() gin.HandlerFunc{
// 	return func(c *gin.Context){
// 		var product models.Product
// 		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
// 		productId := c.Param("id")

// 		objID, _ := primitive.ObjectIDFromHex(productId)

// 		if err := c.BindJSON(&product); err !=nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		validationErr := validate.Struct(product)
// 		if validationErr != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
// 			return
// 		}

// 		product.Updated_at = time.Now()

// 		update := bson.M{
// 			"name":			product.Name,
// 			"description":	product.Description,
// 			"price":		product.Price,
// 			"image":		product.Image,
// 			"updated_at":	product.Updated_at,
// 			"user_id":		product.UserID,
// 		}

// 		resutl, err := productCollection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
// 		if err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}
		
// 		if resutl.MatchedCount == 0{
// 			c.JSON(http.StatusNotFound, gin.H{"error":"Product not found"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Product update successful"})
// 		defer cancel()
// 	}
// }

func EditProduct() gin.HandlerFunc{
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		productID := c.Param("id")

		objID, err := primitive.ObjectIDFromHex(productID)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"errpr": "Invalid product"})
			return
		}

		form, err := c.MultipartForm()
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
			return
		}

		var product models.Product
		productData := form.Value["product"]
		if len(productData) > 0{
			if err := json.Unmarshal([]byte(productData[0]), &product); err != nil{
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product data form"})
				return
			}
		}

		validationErr := validate.Struct(product)
		if validationErr != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		if fileHeaders, ok := form.File["image"]; ok && len(fileHeaders) > 0 {
			fileHeader := fileHeaders[0]
			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open file"})
				return
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
				return
			}

			// Tạo bucket GridFS
			bucket, err := gridfs.NewBucket(
				productCollection.Database(),
				options.GridFSBucket().SetName("images"),
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create GridFS bucket"})
				return
			}

			// Tạo upload stream
			uploadStream, err := bucket.OpenUploadStream(fileHeader.Filename)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload stream"})
				return
			}
			defer uploadStream.Close()

			// Ghi dữ liệu vào GridFS
			if _, err := uploadStream.Write(data); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file to GridFS"})
				return
			}

			// Lấy fileID từ uploadStream
			fileID, ok := uploadStream.FileID.(primitive.ObjectID)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert file ID to ObjectID"})
				return
			}

			// Cập nhật ảnh trong sản phẩm
			product.Image = fileID.Hex()
		}

		product.Updated_at = time.Now()
		update := bson.M{
			"name":				product.Name,
			"description":		product.Description,
			"price":			product.Price,
			"image":			product.Image,
			"updated_at":		product.Updated_at,
			"user_id":			product.UserID,
		}

		result, err := productCollection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if result.MatchedCount == 0{
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product updated successful"})
		defer cancel()
	}
}

func DeleteProduct() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		productID := c.Param("id")

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

func GetAllProducts() gin.HandlerFunc{
	return func (c *gin.Context)  {
		var products []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		cursor, err := productCollection.Find(ctx, bson.M{})
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load products"})
			return
		}

		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &products); err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode data"})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, products)

	}
}