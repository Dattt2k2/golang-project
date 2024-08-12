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
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cartCollection *mongo.Collection = database.OpenCollection(database.Client, "cart")


func AddToCart() gin.HandlerFunc{
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var cartItem models.CartItem

		if err := c.ShouldBindJSON(&cartItem); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		UserID, err := primitive.ObjectIDFromHex(c.Param("userId"))
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
			return
		}

		update := bson.M{
			"$push": bson.M{
				"items": bson.M{
					"$each": []models.CartItem{cartItem},
					"$position": 0,
				},
			},
			"$set": bson.M{"updated_at": time.Now()},
		}
		opts := options.Update().SetUpsert(true)
		_, err = cartCollection.UpdateOne(ctx, bson.M{"user_id": UserID},update, opts)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add items to cart"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Item added to cart successfully"})
		defer cancel()
	}
}

func GetCart() gin.HandlerFunc{
	return func(c *gin.Context) {
		
	}
}