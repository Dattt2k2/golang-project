package controller

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Dattt2k2/golang-project/order-service/database"
	"github.com/Dattt2k2/golang-project/order-service/kafka"
	"github.com/Dattt2k2/golang-project/order-service/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	// "google.golang.org/grpc"

	cartpb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
    services "github.com/Dattt2k2/golang-project/order-service/service"
)

// func OrderFromCart() gin.HandlerFunc{
// 	return func(c *gin.Context){

// 	}
// }

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

func CheckUserRole(c *gin.Context) {
	userRole := c.GetHeader("user_type")
	if userRole != "USER" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You don't have permission"})
		c.Abort()
        return
	}
    c.Next()
}


func OrderFromCart() gin.HandlerFunc{
	return func(c *gin.Context){
		

		CheckUserRole(c)
        if c.IsAborted(){
            return
        }

		// userIdStr := c.Param("userId")
		// userId, err := primitive.ObjectIDFromHex(userIdStr)
		// if err != nil{
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	return 
		// }

        uid := c.GetHeader("user_id")

		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get userID"})
			return
		}
        

		client := services.CartServiceConnection()
        if client  == nil{
            c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Cart service unavailable"})
            return
        }
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req:= &cartpb.CartRequest{
			UserId: userID.Hex(),
		}

		resp, err := client.GetCartItems(ctx, req)
		if err != nil{
			log.Printf("Failed to get cart items: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var orderItems []models.OrderItem
		var totalPrice float64 = 0

		for _, item:= range resp.Items{
			productId, err := primitive.ObjectIDFromHex(item.ProductId)
			if err != nil{
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			orderItem := models.OrderItem{
				ProductID: productId,
                Name: item.Name,
				Quantity: int(item.Quantity),
				Price : float64(item.Price),
			}

			orderItems = append(orderItems, orderItem)
			totalPrice += float64(item.Quantity) * float64(item.Price)
		}

		now := time.Now()
		newOrder:= models.Order{
			ID: primitive.NewObjectID(),
			UserID: userID,
			Items: orderItems,
			TotalPrice: totalPrice,
			Status: "PENDING",
			Source: "CART",
			Created_at: now,
			Updated_at: now,
		}

		_, err = orderCollection.InsertOne(ctx, newOrder)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		orderEvent:= kafka.PaymentOrder{
			UserId: uid,
			Amount: totalPrice,
			Products: resp.Items,
		}

		if err := kafka.ProducePaymentOrder(orderEvent); err != nil{
			log.Printf("Failed to produce payment order: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully",
			"order_id": newOrder.ID.Hex(),
			"total_price": totalPrice,
		})

	}
}

func OrderDirectly() gin.HandlerFunc {
    type ProductRequest struct {
        ProductID string  `json:"product_id" binding:"required"`
        Name      string  `json:"name" binding:"required"`
        Quantity  int     `json:"quantity" binding:"required"`
        Price     float64 `json:"price" binding:"required"`
    }

    type OrderRequest struct {
        UserID  string           `json:"user_id" binding:"required"`
        Items   []ProductRequest `json:"items" binding:"required,dive"`
        Source  string           `json:"source" binding:"required"`
    }

    return func(c *gin.Context) {
        CheckUserRole(c)

        var orderReq OrderRequest
        if err := c.ShouldBindJSON(&orderReq); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Convert user ID from string to ObjectID
        userId, err := primitive.ObjectIDFromHex(orderReq.UserID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
            return
        }

        var orderItems []models.OrderItem
        var totalPrice float64 = 0

        // Convert product request items to order items
        for _, item := range orderReq.Items {
            productId, err := primitive.ObjectIDFromHex(item.ProductID)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
                return
            }

            orderItem := models.OrderItem{
                ProductID: productId,
                Name:      item.Name,
                Quantity:  item.Quantity,
                Price:     item.Price,
            }

            orderItems = append(orderItems, orderItem)
            totalPrice += float64(item.Quantity) * item.Price
        }

        // Create the order
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        now := time.Now()
        newOrder := models.Order{
            ID:          primitive.NewObjectID(),
            UserID:      userId,
            Items:       orderItems,
            TotalPrice:  totalPrice,
            Status:      "pending",
            Source:      orderReq.Source,
            Created_at:  now,
            Updated_at:  now,
        }

        // Save to MongoDB
        _, err = orderCollection.InsertOne(ctx, newOrder)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
            return
        }

        // Send to kafka for payment processing
        kafkaItems := make([]interface{}, len(orderItems))
        for i, item := range orderItems {
            kafkaItems[i] = map[string]interface{}{
                "product_id": item.ProductID.Hex(),
                "name":       item.Name,
                "quantity":   item.Quantity,
                "price":      item.Price,
            }
        }

        orderEvent := kafka.PaymentOrder{
            UserId:   orderReq.UserID,
            Amount:   totalPrice,
            Products: kafkaItems,
        }

        if err := kafka.ProducePaymentOrder(orderEvent); err != nil {
            log.Printf("Failed to produce payment order: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Order created but payment processing failed"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Order placed successfully",
            "order_id": newOrder.ID.Hex(),
            "total": totalPrice,
        })
    }
}

func calculateAmount(resp *cartpb.CartResponse) float64 {
    var total float64
    for _, item := range resp.Items {
        total += float64(item.Quantity) * float64(item.Price)
    }
    return total
}

// func OrderFromProduct() gin.HandlerFunc{
// 	return func(c *gin.Context){
// 		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
// 		if err != nil{
// 			log.Printf("Failed to connect to gRPC server: %v", err)
// 			return
// 		}
// 		defer conn.Close()

// 		CheckUserRole(c)

// 		client:= cartpb.NewCartServiceClient(conn)
// 		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		defer cancel()

// 		req:= &cartpb.CartRequest{
// 			UserId: c.Param("userId"),
// 		}

// 		resp, err := client.GetCartItems(ctx, req)
// 		if err != nil{
// 			log.Printf("Failed to get cart items: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

// 		total := calculateAmount(resp)
// 		log.Printf("Total amount: %v", total)

// 		orderEvent:= kafka.PaymentOrder{
// 			UserId: req.UserId,
// 			Amount: total,
// 			Products: resp.Items,
// 		}

// 		if err := kafka.ProducePaymentOrder(orderEvent); err != nil{
// 			log.Printf("Failed to produce payment order: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}


// 		c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully"})

// 	}

// }

// func OrderDirectly() gin.HandlerFunc{

// 	type OrderRequest struct{
// 		Userid string `json:"user_id"`
// 		Items []struct{
// 			ProductId string `json:"product_id"`
// 			Quantity int32 `json:"quantity"`
// 			Price float64 `json:"price"`
// 		} `json:"items"`
// 	}

// 	return func(c *gin.Context){
		
// 		CheckUserRole(c)

// 		var orderReq OrderRequest
// 		if err := c.ShouldBindJSON(&orderReq); err != nil{
// 			log.Printf("Failed to bind JSON: %v", err)
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		var total float64
// 		for _, items := range orderReq.Items{
// 			total += float64(items.Quantity) * float64(items.Price)
// 		}

// 		orderEvent:= kafka.PaymentOrder{
// 			UserId: orderReq.Userid,
// 			Amount: total,
// 			Products: orderReq.Items,
// 		}

// 		if err := kafka.ProducePaymentOrder(orderEvent); err != nil{
// 			log.Printf("Failed to produce payment order: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully"})
// 	}


// }

// func calculateAmount(resp *cartpb.CartResponse) float64{
// 	var total float64
// 	for _, item := range resp.Items{
// 		total += float64(item.Quantity) * float64(item.Price)
// 	}
// 	return total
// }