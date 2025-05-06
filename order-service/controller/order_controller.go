// package controller

// import (
// 	"context"
// 	"log"
// 	"math"
// 	"net/http"
// 	"strconv"
// 	"time"

// 	"github.com/Dattt2k2/golang-project/order-service/database"
// 	"github.com/Dattt2k2/golang-project/order-service/models"
// 	"github.com/gin-gonic/gin"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
//     "github.com/Dattt2k2/golang-project/order-service/kafka"
// 	// "google.golang.org/grpc"

// 	cartpb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
// 	services "github.com/Dattt2k2/golang-project/order-service/service"
// )

// // func OrderFromCart() gin.HandlerFunc{
// // 	return func(c *gin.Context){

// // 	}
// // }

// var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

// func CheckUserRole(c *gin.Context) {
// 	userRole := c.GetHeader("user_type")
// 	if userRole != "USER" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "You don't have permission"})
// 		c.Abort()
//         return
// 	}
//     c.Next()
// }

// func OrderFromCart() gin.HandlerFunc{
// 	return func(c *gin.Context){

// 		CheckUserRole(c)
//         if c.IsAborted(){
//             return
//         }

// 		// userIdStr := c.Param("userId")
// 		// userId, err := primitive.ObjectIDFromHex(userIdStr)
// 		// if err != nil{
// 		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		// 	return
// 		// }

//         uid := c.GetHeader("user_id")

// 		userID, err := primitive.ObjectIDFromHex(uid)
// 		if err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get userID"})
// 			return
// 		}

// 		client := services.CartServiceConnection()
//         if client  == nil{
//             c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Cart service unavailable"})
//             return
//         }
// 		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		defer cancel()

// 		req:= &cartpb.CartRequest{
// 			UserId: userID.Hex(),
// 		}

// 		resp, err := client.GetCartItems(ctx, req)
// 		if err != nil{
// 			log.Printf("Failed to get cart items: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

// 		var orderItems []models.OrderItem
// 		var totalPrice float64 = 0

// 		for _, item:= range resp.Items{
//             productId, err := primitive.ObjectIDFromHex(item.ProductId)
//             if err != nil{
//                 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//                 return
//             }

// 			orderItem := models.OrderItem{
// 				ProductID: productId,
//                 Name: item.Name,
// 				Quantity: int(item.Quantity),
// 				Price : float64(item.Price),
// 			}

// 			orderItems = append(orderItems, orderItem)
// 			totalPrice += float64(item.Quantity) * float64(item.Price)
// 		}

// 		now := time.Now()
// 		newOrder:= models.Order{
// 			ID: primitive.NewObjectID(),
// 			UserID: userID,
// 			Items: orderItems,
// 			TotalPrice: totalPrice,
// 			Status: "PENDING",
// 			Source: "CART",
// 			Created_at: now,
// 			Updated_at: now,
// 		}

// 		_, err = orderCollection.InsertOne(ctx, newOrder)
// 		if err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

//         if err := kafka.ProduceOrderSuccessEvent(ctx, newOrder); err != nil{
//             log.Printf("warning: Failed to produce order created event: %v", err)
//         }

//         c.JSON(http.StatusOK, gin.H{
//             "message": "Order placed successfully",
//             "order_id": newOrder.ID.Hex(),
//             "total": totalPrice,
//         })
// 	}
// }

// func OrderDirectly() gin.HandlerFunc {
//     type ProductRequest struct {
//         ProductID string  `json:"product_id" binding:"required"`
//         Name      string  `json:"name" binding:"required"`
//         Quantity  int     `json:"quantity" binding:"required"`
//         Price     float64 `json:"price" binding:"required"`
//     }

//     type OrderRequest struct {
//         UserID  string           `json:"user_id" binding:"required"`
//         Items   []ProductRequest `json:"items" binding:"required,dive"`
//         Source  string           `json:"source" binding:"required"`
//     }

//     return func(c *gin.Context) {
//         CheckUserRole(c)

//         var orderReq OrderRequest
//         if err := c.ShouldBindJSON(&orderReq); err != nil {
//             c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//             return
//         }

//         // Convert user ID from string to ObjectID
//         userId, err := primitive.ObjectIDFromHex(orderReq.UserID)
//         if err != nil {
//             c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
//             return
//         }

//         var orderItems []models.OrderItem
//         var totalPrice float64 = 0

//         // Convert product request items to order items
//         for _, item := range orderReq.Items {
//             productId, err := primitive.ObjectIDFromHex(item.ProductID)
//             if err != nil {
//                 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
//                 return
//             }

//             orderItem := models.OrderItem{
//                 ProductID: productId,
//                 Name:      item.Name,
//                 Quantity:  item.Quantity,
//                 Price:     item.Price,
//             }

//             orderItems = append(orderItems, orderItem)
//             totalPrice += float64(item.Quantity) * item.Price
//         }

//         // Create the order
//         ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//         defer cancel()

//         paymentMethod := c.Query("payment_method")
//         shippingAddress := c.Query("shipping_address")
//         if paymentMethod != "COD" && paymentMethod != "ONLINE" {
//             c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method"})
//             return
//         }
//         if paymentMethod == "" {
//             paymentMethod = "COD" // Default to COD
//         }
//         if shippingAddress == "" {
//             c.JSON(http.StatusBadRequest, gin.H{"error": "Shipping address is required"})
//             return
//         }

//         initialStatus := "PENDING"
//         paymentStatus := "PENDING"

//         if paymentMethod == "COD" {
//             initialStatus = "PROCESSING"
//             paymentStatus = "PENDING_VERIFICATION"

//         }

//         now := time.Now()
//         newOrder := models.Order{
//             ID:          primitive.NewObjectID(),
//             UserID:      userId,
//             Items:       orderItems,
//             TotalPrice:  totalPrice,
//             Status:      initialStatus,
//             PaymentMethod: paymentMethod,
//             PaymentStatus: paymentStatus,
//             ShippingAddress: shippingAddress,
//             Source:      "CART",
//             Created_at:  now,
//             Updated_at:  now,
//         }

//         // Save to MongoDB
//         _, err = orderCollection.InsertOne(ctx, newOrder)
//         if err != nil {
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
//             return
//         }

//         if err := kafka.ProduceOrderSuccessEvent(ctx, newOrder); err != nil {
//             log.Printf("warning: Failed to produce order created event: %v", err)
//         }

//         // Send to kafka for payment processing
//         kafkaItems := make([]interface{}, len(orderItems))
//         for i, item := range orderItems {
//             kafkaItems[i] = map[string]interface{}{
//                 "product_id": item.ProductID.Hex(),
//                 "name":       item.Name,
//                 "quantity":   item.Quantity,
//                 "price":      item.Price,
//             }
//         }

//         c.JSON(http.StatusOK, gin.H{
//             "message": "Order placed successfully",
//             "order_id": newOrder.ID.Hex(),
//             "total": totalPrice,
//         })
//     }
// }

// func calculateAmount(resp *cartpb.CartResponse) float64 {
//     var total float64
//     for _, item := range resp.Items {
//         total += float64(item.Quantity) * float64(item.Price)
//     }
//     return total
// }

// func GetOrder() gin.HandlerFunc{
//     return func (c *gin.Context){
//         ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//         defer cancel()

//         CheckUserRole(c)
//         if c.IsAborted(){
//             return
//         }

//         var orders []models.Order

//         page, err := strconv.Atoi(c.Query("page"))
//         if err != nil{
//             log.Printf("Failed to parse page: %v", err)
//             page = 1
//         }
//         limit, err := strconv.Atoi(c.Query("limit"))
//         if err != nil{
//             log.Printf("Failed to parse limit: %v", err)
//             limit = 10
//         }

//         log.Printf("Pagination: page %d, limit %d", page, limit)

//         skip := (page - 1) * limit

//         total, err := orderCollection.CountDocuments(ctx, bson.M{})
//         if err != nil{
//             log.Printf("Failed to count documents: %v", err)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count documents"})
//             return
//         }

//         log.Printf("Total documents: %d", total)

//         if total == 0{
//             c.JSON(http.StatusOK, gin.H{
//                 "data": []models.Order{},
//                 "total": 0,
//                 "page": page,
//                 "limit": limit,
//                 "pages": 0,
//                 "has_next": false,
//                 "has_prev": false,
//             })
//             return
//         }

//         findOptions := options.Find().
//             SetSkip(int64(skip)).
//             SetLimit(int64(limit)).
//             SetSort(bson.D{{Key:"created_at",Value: -1}})

//         cursor, err := orderCollection.Find(ctx, bson.M{}, findOptions)
//         if err != nil{
//             log.Printf("Failed to find documents: %v", err)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fect order"})
//             return
//         }
//         defer cursor.Close(ctx)

//         if err := cursor.All(ctx, &orders); err != nil{
//             log.Printf("Failed to decode documents: %v", err)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode documents"})
//             return
//         }

//         for i := range orders{
//             var items []models.OrderItem
//             itemCursor, err := orderCollection.Aggregate(ctx, mongo.Pipeline{
//                 bson.D{{Key: "$match", Value: bson.M{"_id": orders[i].ID}}},
//                 bson.D{{Key: "unwind", Value: bson.M{"path": "$items"}}},
//                 bson.D{{Key: "$lookup", Value: bson.M{
//                     "product_id": "$items.product_id",
//                     "quantity": "$items.quantity",
//                     "price": "$items.price",
//                     "name": "$items.name",
//                     "image_url": "$items.image_url",
//                     "description": "$items.description",
//                 }}},
//             })
//             if err != nil{
//                 log.Printf("Error fetching order items: %v", err)
//                 continue
//             }
//             defer itemCursor.Close(ctx)

//             if err := itemCursor.All(ctx, &orders); err != nil{
//                 log.Printf("Failed to decode order items: %v", err)
//                 continue
//             }
//             orders[i].Items = items
//         }

//         pages := int(math.Ceil(float64(total)/ float64(limit)))

//         response := gin.H{
//             "data": orders,
//             "total": total,
//             "page": page,
//             "limit": limit,
//             "pages": pages,
//             "has_next": page < pages,
//             "has_prev": page > 1,
//         }

//         log.Printf("Response: %v", response)
//         c.JSON(http.StatusOK, response)
//     }
// }

package controller

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Dattt2k2/golang-project/order-service/log"
	"github.com/Dattt2k2/golang-project/order-service/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


type OrderController struct {
	orderService *service.OrderService
}

func NewOrderController(orderService *service.OrderService)  *OrderController{
	return &OrderController{
		orderService: orderService,
	}
}

// func CheckUserRole(c *gin.Context){
// 	userRole := c.GetHeader("user_type")
// 	if userRole != "USER"{
//         logger.Err("Unauthorized access", nil, logger.Str("user_role", userRole))
// 		c.JSON(http.StatusUnauthorized, gin.H{"error":"Your don't have permission"})
// 		c.Abort()
// 		return
// 	}
// 	c.Next()
// }

// func CheckSellerRole(c *gin.Context){
//     userRole := c.GetHeader("user_type")
//     if userRole != "SELLER"{
//         logger.Err("Unauthorized access", nil, logger.Str("user_role", userRole))
//         c.JSON(http.StatusUnauthorized, gin.H{"error": "Your don't have permission"})
//         c.Abort()
//         return
//     }
// }

func (ctrl *OrderController) OrderFromCart() gin.HandlerFunc{
	return func(c *gin.Context){
		// CheckUserRole(c)
		if c.IsAborted(){
			return
		}

		uid := c.GetHeader("user_id")
		UserID, err := primitive.ObjectIDFromHex(uid)
		if err != nil{
            logger.Err("Failed to parse userID", err, logger.Str("user_id", uid))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
			return
		}

        type OrderCartRequest struct{
            Source string `json:"source"`
            PaymentMethod string `json:"payment_method"`
            ShippingAddress string `json:"shipping_address"`
            SelectedProductIDs []string `json:"selected_product_ids"`
        }

        var requestBody OrderCartRequest
        if err := c.ShouldBindJSON(&requestBody); err != nil{
            logger.Err("Failed to bind JSON", err, logger.Str("request_body", requestBody.Source))
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
            return
        }
        if requestBody.PaymentMethod != "COD" && requestBody.PaymentMethod != "ONLINE" {
            logger.Err("Invalid payment method", nil, logger.Str("payment_method", requestBody.PaymentMethod))
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method"})
            return
        }
        if requestBody.PaymentMethod == "" {
            requestBody.PaymentMethod = "COD"
        }
        if requestBody.ShippingAddress == "" {
            logger.Err("Shipping address is nil", nil, logger.Str("shipping_address", requestBody.ShippingAddress))
            c.JSON(http.StatusBadRequest, gin.H{"error": "Shipping address is required"})
            return
        }

		ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
		defer cancel()

		order, err := ctrl.orderService.CreateOrderFromCart(ctx, UserID, requestBody.Source, requestBody.PaymentMethod, requestBody.ShippingAddress, requestBody.SelectedProductIDs)

		if err != nil{
			if err == service.ErrCartServiceUnavailable {
                logger.Err("Cart service unavailable", err, logger.Str("user_id", uid))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Cart service unavailable"})
				return
			} else{
                logger.Err("Failed to create order", err, logger.Str("user_id", uid))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Order placed successfully",
			"order_id": order.ID.Hex(),
			"total_price": order.TotalPrice,
            "payment_method": order.PaymentMethod,
            "shipping_address": order.ShippingAddress,
            "status": order.Status,
		})

        logger.Info("Order placed successfully", logger.Str("order_id", order.ID.Hex()), logger.Str("user_id", uid), logger.Int("total_price", int(order.TotalPrice)))

	}
}

// Order directly from product, using product ID and quantity
// This function is used when user want to order directly from product page
func (ctrl *OrderController) OrderDirectly() gin.HandlerFunc{
    return func (c *gin.Context){
        // CheckUserRole(c)
        if c.IsAborted(){
            return
        }

        userID := c.GetHeader("user_id")
        if userID == ""{
            logger.Err("Failed to get userID", nil, logger.Str("user_id", userID))
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
            return
        }

        var orderReq service.OrderDirectRequest
        if err := c.ShouldBindJSON(&orderReq); err != nil{
            logger.Err("Failed to bind JSON", err, logger.Str("request_body", orderReq.Source))
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        orderReq.UserID = userID
        ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
        defer cancel()

        order, err :=  ctrl.orderService.CreateOrderDirect(ctx, orderReq)
        if err != nil{
            logger.Err("Failed to create order", err, logger.Str("user_id", userID))
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
            return 
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Order placed successfully",
            "order_id": order.ID.Hex(),
            "total_price": order.TotalPrice,
            "payment_method": order.PaymentMethod,
            "shipping_address": order.ShippingAddress,
            "status": order.Status,
        })

        logger.Info("Order placed successfully", logger.Str("order_id", order.ID.Hex()), logger.Str("user_id", userID), logger.Int("total_price", int(order.TotalPrice)))
    }
}
func (ctrl *OrderController) AdminGetOrders() gin.HandlerFunc{
    return func(c *gin.Context){
        // CheckSellerRole(c)
        if c.IsAborted(){
            return
        }

        page, err := strconv.Atoi(c.Query("page"))
        if err != nil{
            logger.Err("Failed to parse page", err, logger.Str("page", c.Query("page")))
            page = 1
        }

        limit, err := strconv.Atoi(c.Query("limit"))
        if err != nil{
            logger.Err("Failed to parse limit", err, logger.Str("limit", c.Query("limit")))
            limit = 10
        }

        ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
        defer cancel()

        orders, total, pages, hasNext, hasPrev, err := ctrl.orderService.AdminGetOrders(ctx, page, limit)
        if err != nil{
            logger.Err("Failed to get orders", err, logger.Str("page", strconv.Itoa(page)), logger.Str("limit", strconv.Itoa(limit)))
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders"})
            return
        }

        if len(orders) == 0{
            c.JSON(http.StatusOK, gin.H{
                "data": []interface{}{},
                "total": 0,
                "page": page,
                "limit": limit,
                "pages": 0,
                "has_next": false,
                "has_prev": false,
            })
            return 
        }

        c.JSON(http.StatusOK, gin.H{
            "data": orders,
            "total": total,
            "page": page,
            "limit": limit,
            "pages": pages,
            "has_next": hasNext,
            "has_prev": hasPrev,
        })

        logger.Info("Admin get orders successfully", logger.Str("page", strconv.Itoa(page)), logger.Str("limit", strconv.Itoa(limit)))
    }
}

// GetUserOrders retrieves orders for a specific user with pagination
func (ctrl *OrderController) GetUserOrders() gin.HandlerFunc{
    return func (c *gin.Context){
        // CheckUserRole(c)
        if c.IsAborted(){
            return
        }

        uid := c.GetHeader("user_id")
        userID, err := primitive.ObjectIDFromHex(uid)
        if err != nil{
            logger.Err("Failed to parse userID", err, logger.Str("user_id", uid))
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
            return
        }

        page := 1
        limit := 10

        pageStr := c.Query("page")
        if pageStr != ""{
            if p, err := strconv.Atoi(pageStr); err == nil{
                page = p
            }
        }

        limitStr := c.Query("limit")
        if limitStr != ""{
            if l, err := strconv.Atoi(limitStr); err == nil{
                limit = l
            }
        }

        ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
        defer cancel()

        orders, total, pages, hasNext, hasPrev, err := ctrl.orderService.GetUserOrders(ctx, userID, page, limit)
        if err != nil{
            logger.Err("Failed to get orders", err, logger.Str("user_id", uid), logger.Str("page", strconv.Itoa(page)), logger.Str("limit", strconv.Itoa(limit)))
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders"})
            return 
        }

        if len(orders) == 0{
            c.JSON(http.StatusOK, gin.H{
                "data": []interface{}{},
                "total": 0,
                "page": page,
                "limit": limit,
                "pages": 0,
                "has_next": false,
                "has_prev": false,
            })
            return 
        }

        c.JSON(http.StatusOK, gin.H{
            "data": orders,
            "total": total,
            "page": page,
            "limit": limit,
            "pages": pages,
            "has_next": hasNext,
            "has_prev": hasPrev,
        })

        logger.Info("Get user orders successfully", logger.Str("user_id", uid), logger.Str("page", strconv.Itoa(page)), logger.Str("limit", strconv.Itoa(limit)))
    }
}


// Cancel Order with ID
func (ctrl *OrderController) CancelOrder() gin.HandlerFunc{
    return func (c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
        defer cancel()
        userRole := c.GetHeader("user_type")
        if userRole != "USER" && userRole != "SELLER" {
            logger.Err("Unauthorized access", nil, logger.Str("user_role", userRole))
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user type"})
            return
        }
        orderIDStr := c.Param("order_id")
        orderID, err := primitive.ObjectIDFromHex(orderIDStr)
        if err != nil{
            logger.Err("Failed to parse orderID", err, logger.Str("order_id", orderIDStr))
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid orderID"})
            return
        }
        var userID primitive.ObjectID
        if userRole == "USER"{
            userIdHeader := c.GetHeader("user_id")

            if userIdHeader == ""{
                logger.Err("Invalid User ID", nil, logger.Str("user_id", userIdHeader))
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
                return 
            }

            userID, err = primitive.ObjectIDFromHex(userIdHeader)
            if err != nil{
                logger.Err("Failed to parse userID", err, logger.Str("user_id", userIdHeader))
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
                return
            }
        }
        

        err = ctrl.orderService.CanceldOrder(ctx, orderID, userID, userRole)
        if err != nil{
            logger.Err("Failed to cancel order", err, logger.Str("order_id", orderIDStr), logger.Str("user_id", userID.Hex()))
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return 
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Order cancelled successfully"})

        logger.Info("Order cancelled successfully", logger.Str("order_id", orderIDStr), logger.Str("user_id", userID.Hex()))
    }

}


