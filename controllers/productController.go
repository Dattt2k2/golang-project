package controllers

import (
	"context"
	"net/http"
	"time"

	database "github.com/Dattt2k2/golang-project/database/databaseConnection.gp"
	"github.com/Dattt2k2/golang-project/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)


var productCollection *mongo.Collection = database.OpenCollection(database.Client, "product")

func AddProduct() gin.HandlerFunc{
	return func(c *gin.Context){
		var product models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		if err := c.BindJSON(&product); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validate.Struct(product)
		if validationErr != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		
		product.Created_at = time.Now()
		product.Updated_at = time.Now()
		product.ID = primitive.NewObjectID()

		result, err := productCollection.InsertOne(ctx, product)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": result.InsertedID})
		defer cancel()
	}
}

func EditProduct() gin.HandlerFunc{
	return func(c *gin.Context){
		var product models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		productId := c.Param("id")

		objID, _ := primitive.ObjectIDFromHex(productId)

		if err := c.BindJSON(&product); err !=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(product)
		if validationErr != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		product.Updated_at = time.Now()

		update := bson.M{
			"name":			product.Name,
			"description":	product.Description,
			"price":		product.Price,
			"image":		product.Image,
			"updated_at":	product.Updated_at,
		}

		resutl, err := productCollection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		if resutl.MatchedCount == 0{
			c.JSON(http.StatusNotFound, gin.H{"error":"Product not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product update successful"})
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