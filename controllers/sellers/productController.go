package controllers

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	database "github.com/Dattt2k2/golang-project/database/databaseConnection.gp"
	"github.com/Dattt2k2/golang-project/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
)


var productCollection *mongo.Collection = database.OpenCollection(database.Client, "product")
var validate = validator.New()



func AddProduct() gin.HandlerFunc{
	return func(c *gin.Context) {
		var product models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)



		name := c.PostForm("name")
		description := c.PostForm("description")
		price := c.PostForm("price")

		file, err := c.FormFile("image")
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
			return
		}

		fileContent, err := file.Open()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
			return
		}

		defer fileContent.Close()

		imageBytes, err := io.ReadAll(fileContent)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image content"})
			return
		}

		userID:= c.GetString("uid")
		if userID == ""{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get userID"})
			return
		}

		product.Name = &name
		product.Description = &description
		product.Image = primitive.Binary{Data: imageBytes}
		product.Price = parseFloat(price)
		product.UserID, _ = primitive.ObjectIDFromHex(userID)

		if err := validate.Struct(product); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now()
		product.Created_at = now
		product.Updated_at = now

		product.ID = primitive.NewObjectID()

		_, err = productCollection.InsertOne(ctx, product)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product added successfully", "product": product.Name})
		defer cancel()
	}
}

func parseFloat(s string) float64{
	f, _ := strconv.ParseFloat(s, 64)
	return f
}


func EditProduct() gin.HandlerFunc{
	return func (c *gin.Context)  {
		// var product models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		productID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
			return
		}

		name := c.PostForm("name")
		description := c.PostForm("description")
		price := c.PostForm("price")


		update := bson.M{"updated_at": time.Now()}
		
		if name != ""{
			update["name"] = name
		}

		if description != ""{
			update["description"] = description
		}

		if price != ""{
			update["price"] = parseFloat(price)
		}


		file, err := c.FormFile("image")
		if err == nil{
			fileContent, err := file.Open()
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
				return
			}
			defer fileContent.Close()
			imageBytes, err := io.ReadAll(fileContent)
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image content"})
				return
			}

			update["image"] = primitive.Binary{Data: imageBytes}
		}

		result, err := productCollection.UpdateOne(
			ctx,
			bson.M{"_id": productID},
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

