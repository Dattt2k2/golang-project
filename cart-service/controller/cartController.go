// package controllers

// import (
// 	"context"
// 	"net/http"
// 	"time"

// 	database "github.com/Dattt2k2/golang-project/database/databaseConnection.gp"
// 	"github.com/Dattt2k2/golang-project/cart-service/models"
// 	"github.com/gin-gonic/gin"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// var cartCollection *mongo.Collection = database.OpenCollection(database.Client, "cart")

// func AddToCart() gin.HandlerFunc{
// 	return func(c *gin.Context) {
// 		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

// 		var cartItem models.CartItem

// 		if err := c.ShouldBindJSON(&cartItem); err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		UserID, err := primitive.ObjectIDFromHex(c.Param("userId"))
// 		if err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
// 			return
// 		}

// 		update := bson.M{
// 			"$push": bson.M{
// 				"items": bson.M{
// 					"$each": []models.CartItem{cartItem},
// 					"$position": 0,
// 				},
// 			},
// 			"$set": bson.M{"updated_at": time.Now()},
// 		}
// 		opts := options.Update().SetUpsert(true)
// 		_, err = cartCollection.UpdateOne(ctx, bson.M{"user_id": UserID},update, opts)
// 		if err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add items to cart"})
// 			return
// 		}
// 		c.JSON(http.StatusOK, gin.H{"message": "Item added to cart successfully"})
// 		defer cancel()
// 	}
// }

// func GetCart() gin.HandlerFunc{
// 	return func(c *gin.Context) {
// 		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

// 		var cartItem[] models.CartItem

// 		cursor, err := cartCollection.Find(ctx, bson.M{})
// 		if err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to load cart"})
// 			return
// 		}

// 		defer cursor.Close(ctx)

// 		if err := cursor.All(ctx, &cartItem); err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load cart"})
// 			return
// 		}
// 		defer cancel()

// 		c.JSON(http.StatusOK, cartItem)
// 	}
// }

// func DeleteProductFromCart() gin.HandlerFunc{
// 	return func (c *gin.Context)  {
// 		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

// 		productId := c.Param("id")

// 		objId, _ := primitive.ObjectIDFromHex(productId)

// 		result, err := cartCollection.DeleteOne(ctx, bson.M{"_id": objId})
// 		if err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete product from cart"})
// 			return
// 		}
// 		if result.DeletedCount == 0{
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Delete product from cart successful"})
// 		defer cancel()

// 	}
// }

// func GetProductFromCart() gin.HandlerFunc{
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

// 		var product []models.Product
// 		productId := c.Param("id")
// 		objId, _ := primitive.ObjectIDFromHex(productId)

// 		filter := bson.D{{"_id", objId}}

// 		cursor, err := productCollection.Find(ctx, filter)

// 		if err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product"});
// 			return
// 		}
// 		defer cursor.Close(ctx)

// 		if err := cursor.All(ctx, &product); err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode data"})
// 			return
// 		}
// 		defer cancel()

// 		c.JSON(http.StatusOK, product)
// 	}
// }

package controller

import (
	// "container/list"
	"context"
	"log"
	"net/http"
	// "os/user"
	"strconv"
	"time"

	"github.com/Dattt2k2/golang-project/cart-service/database"
	"github.com/Dattt2k2/golang-project/cart-service/models"
	pb "github.com/Dattt2k2/golang-project/module/gRPC-Product/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var cartCollection *mongo.Collection = database.OpenCollection(database.Client, "cart")
var productClient pb.ProductServiceClient

func  InitProductServiceConnection(){
	conn, err := grpc.NewClient("product-service:8089", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil{
		log.Fatalf("Coulnd not connect to product service: %v", err)
	}

	log.Printf("Connect to product-service")
	productClient = pb.NewProductServiceClient(conn)
}



func CheckUserRole(c *gin.Context) {
	userRole := c.GetHeader("user_type")
	log.Println(userRole)
	if userRole != "USER" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You don't have permission"})
		c.Abort()
		return
	}
}


func AddToCart() gin.HandlerFunc{
	return func(c *gin.Context){
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		CheckUserRole(c)
		if c.IsAborted(){
			return
		}

		productId := c.Param("id")

		if productId == ""{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product id not found"})
			return
		}

		uid := c.GetHeader("user_id")

		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get userID"})
			return
		}

	
		productReq := &pb.ProductRequest{Id : productId}
		basicInfo, err := productClient.GetBasicInfo(ctx, productReq)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product info"})
			return
		}
		checkStock, err := productClient.CheckStock(ctx, productReq)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check stock"})
			return
		}

		productiontId, err := primitive.ObjectIDFromHex(productId)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Can not convert id to objectID"})
			return
		}
		cartItem := models.CartItem{
			ProductID: productiontId,
			Name: basicInfo.Name,
			Price: float64(basicInfo.Price),
			Quantity: int(checkStock.AvailableQuantity),
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

		opt := options.Update().SetUpsert(true)
		_, err = cartCollection.UpdateOne(ctx, bson.M{"user_id": userID }, update, opt)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product to cart"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Item added to cart successfully"})
	}
}


// Get ID from url and get data from product-service with gRPC
// func AddToCart() gin.HandlerFunc{
// 	return func(c *gin.Context){
// 		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
// 		defer cancel()

// 		// Get id from url
// 		productId := c.Param("id")
// 		if productId == ""{
// 			log.Printf("Product id not found")
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Product id not found"})
// 			return
// 		}

// 		// Convert id to ObjectID
// 		objectID, err := primitive.ObjectIDFromHex(productId)
// 		if err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
// 			return
// 		}

// 		log.Printf("Product ID: %v", objectID)

// 		var cartItem models.CartItem

// 		userID := c.GetHeader("uid")
// 		CheckUserRole(c)

// 		productReq := &pb.ProductRequest{Id: objectID.Hex()}
// 		productRes, err := productionClient.GetProductInfo(ctx, productReq)
// 		log.Printf(c.Request.Proto)
// 		if err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product data"})
// 			return
// 		}

// 		cartItem.Name = productRes.Name
// 		cartItem.Price = float64(productRes.Price)
// 		cartItem.Quantity = int(productRes.Quantity)


// 		update := bson.M{
// 			"$push": bson.M{
// 				"items": bson.M{
// 					"$each": []models.CartItem{cartItem},
// 					"$position": 0,
// 				},
// 			},
// 			"$set": bson.M{"updated_at": time.Now()},
// 		}

// 		opt := options.Update().SetUpsert(true)
// 		_, err = cartCollection.UpdateOne(ctx, bson.M{"user_id": userID}, update, opt)
// 		if err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product to cart"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Item added to cart successfully"})
// 	}
// }


func  GetProductFromCart() gin.HandlerFunc{
	return func(c *gin.Context){
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID := c.GetHeader("user_id")
		if userID == ""{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get userID"})
			return
		}

		CheckUserRole(c)
		if c.IsAborted(){
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if page <= 0{
			page = 1
		}
		if limit <= 0{
			limit = 10
		}

		skip := (page - 1) *limit

		cursor, err := cartCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{"$unwind", "$items"}},
			bson.D{{"$skip", skip}},
			bson.D{{"$limit", limit}},
		})

		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get  product"})
			return
		}

		defer cursor.Close(ctx)


		var cartItems []models.CartItem

		if err := cursor.All(ctx, &cartItems); err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode data "})
			return
		}

		var products []models.CartItem

		for _, item := range cartItems{
			productReq := &pb.ProductRequest{Id: item.ProductID.Hex()}
			productRes, err := productClient.GetProductInfo(ctx, productReq)
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product data"})
				return
			}

			item.Name = productRes.Name
			item.Price = float64(productRes.Price)
			item.Quantity = int(productRes.Quantity)
			item.ImageUrl = productRes.GetImageUrl()
			products = append(products, item)
		}

		c.JSON(http.StatusOK, gin.H{
			"page": page,
			"limit": limit,
			"products": products,
		})
	}
}

// func GetProductFromCart() gin.HandlerFunc{
// 	return func(c *gin.Context){
// 		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
// 		defer cancel()

// 		userID := c.GetHeader("uid")
// 		if userID == ""{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
// 			return
// 		}

// 		CheckUserRole(c)


// 		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
// 		limit, _ := strconv.Atoi(c.DefaultQuery("limit","10"))

// 		if page <= 0{
// 			page = 1
// 		}
// 		if limit <= 0{
// 			limit = 10
// 		}

// 		skip := (page - 1) * limit

// 		cursor, err := cartCollection.Aggregate(ctx, mongo.Pipeline{
// 			bson.D{{"$unwind", "$items"}},
// 			bson.D{{"$skip", skip}},
// 			bson.D{{"$limit", limit}},
// 		})

// 		if err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product"})
// 			return
// 		}

// 		defer cursor.Close(ctx)

// 		var products []models.CartItem
		
// 		if err := cursor.All(ctx, &products); err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode data"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{
// 			"page": page,
// 			"limit": limit,
// 			"products": products,
// 		})

// 	}
// }

func DeleteProductFromCart() gin.HandlerFunc{
	return func (c *gin.Context)  {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID := c.GetHeader("uid")
		if userID == ""{
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authorized"})
			return
		}

		CheckUserRole(c)
		if c.IsAborted(){
			return
		}
		
		var requestBody struct{
			productID []string `bson:"product_id"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get productID"})
			return
		}

		var objectID []primitive.ObjectID

		for _, id := range requestBody.productID{
			objID, err := primitive.ObjectIDFromHex(id)
			if err != nil{
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
				return
			}
			objectID = append(objectID, objID)
		}

		filter := bson.M{
			"items": bson.M{
				"$elementMatch" : bson.M{
					"$product_id": bson.M{"$in":objectID},
				},
			},
		}

		update := bson.M{
			"$pull": bson.M{
				"$items":  bson.M{
					"$product_id": bson.M{"$in":objectID},
				},
			},
		}

		result, err := cartCollection.UpdateOne(ctx, filter, update)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product from cart"})
			return
		}

		if result.ModifiedCount == 0{
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Delete product from cart successfully"})
	}
}