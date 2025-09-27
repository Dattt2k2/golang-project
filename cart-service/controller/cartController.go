// package controller

// import (
// 	// "container/list"
// 	"context"
// 	"log"
// 	"math"
// 	"net/http"

// 	// "os/user"
// 	"strconv"
// 	"time"

// 	"cart-service/database"
// 	"cart-service/models"
// 	pb "module/gRPC-Product/service"
// 	"github.com/gin-gonic/gin"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/credentials/insecure"
// )

// var cartCollection *mongo.Collection = database.OpenCollection(database.Client, "cart")
// var productClient pb.ProductServiceClient

// type CartController struct{
// 	ClearCartFunc func(ctx context.Context, userID primitive.ObjectID)
// }

// func  InitProductServiceConnection(){
// 	conn, err := grpc.NewClient("product-service:8089", grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil{
// 		log.Fatalf("Could not connect to product service: %v", err)
// 	}

// 	log.Printf("Connect to product-service")
// 	productClient = pb.NewProductServiceClient(conn)
// }

// func CheckUserRole(c *gin.Context) {
// 	userRole := c.GetHeader("user_type")
// 	log.Println(userRole)
// 	if userRole != "USER" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "You don't have permission"})
// 		c.Abort()
// 		return
// 	}
// }

// func CheckSellerRole(c *gin.Context){
// 	userRole := c.GetHeader("user_type")
// 	log.Println(userRole)
// 	if userRole != "SELLER" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "You don't have permission"})
// 		c.Abort()
// 		return
// 	}
// }

// func AddToCart() gin.HandlerFunc{
// 	return func(c *gin.Context){
// 		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
// 		defer cancel()

// 		CheckUserRole(c)
// 		if c.IsAborted(){
// 			return
// 		}

// 		productId := c.Param("id")

// 		if productId == ""{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Product id not found"})
// 			return
// 		}

// 		var requestBody struct {
// 			Quantity int `json:"quantity" binding:"required"`
// 		}

// 		if err := c.ShouldBindJSON(&requestBody); err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get quantity"})
// 			return
// 		}

// 		requestedQuantity := requestBody.Quantity

// 		uid := c.GetHeader("user_id")

// 		userID, err := primitive.ObjectIDFromHex(uid)
// 		if err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get userID"})
// 			return
// 		}

// 		productReq := &pb.ProductRequest{Id : productId}
// 		log.Printf("Product id: %v", productId)
// 		basicInfo, err := productClient.GetBasicInfo(ctx, productReq)
// 		if err != nil{
// 			log.Printf(err.Error())
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product info"})
// 			return
// 		}
// 		checkStock, err := productClient.CheckStock(ctx, productReq)
// 		if err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check stock"})
// 			return
// 		}

// 		avaiableQuantity := int(checkStock.AvailableQuantity)

// 		if requestedQuantity > avaiableQuantity {
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"error": "Not enough stock available",
// 				"available_quantity": avaiableQuantity,
// 				"requested_quantity": requestedQuantity,
// 			})
// 			return
// 		}

// 		productiontId, err := primitive.ObjectIDFromHex(productId)
// 		if err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Can not convert id to objectID"})
// 			return
// 		}
// 		cartItem := models.CartItem{
// 			ProductID: productiontId,
// 			Name: basicInfo.Name,
// 			Price: float64(basicInfo.Price),
// 			Quantity: requestedQuantity,
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

// 		opt := options.Update().SetUpsert(true)
// 		_, err = cartCollection.UpdateOne(ctx, bson.M{"user_id": userID }, update, opt)
// 		if err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product to cart"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Item added to cart successfully"})
// 	}
// }

// func GetCartSeller() gin.HandlerFunc {
// 	return func(c *gin.Context){
// 		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
// 		defer cancel()

// 		CheckSellerRole(c)
// 		if c.IsAborted(){
// 			return
// 		}

// 		var carts []models.Cart

// 		page, err :=strconv.Atoi(c.DefaultQuery("page", "1"))
// 		if err != nil{
// 			log.Printf("Invalid page parameter, using default: %v", err)
// 			page = 1
// 		}
// 		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
// 		if err != nil{
// 			log.Printf("Invalid limit parameter, using default: %v", err)
// 			limit = 10
// 		}

// 		log.Printf("Pagination: page=%d, limit=%d", page, limit)

// 		skip := (page - 1) * limit

// 		total, err := cartCollection.CountDocuments(ctx, bson.M{})
// 		if err != nil {
// 			log.Printf("Error counting cart: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count cart documents"})
// 			return
// 		}

// 		log.Printf("Total product count: %d", total)

// 		if total == 0{
// 			c.JSON(http.StatusOK, gin.H{
// 				"data" : []models.Cart{},
// 				"total": 0,
// 				"page": page,
// 				"pages" : 0,
// 				"has_next": false,
// 				"has_prev": false,
// 			})
// 			return
// 		}
// 		findOptions := options.Find().
// 			SetSkip(int64(skip)).
// 			SetLimit(int64(limit)).
// 			SetSort(bson.D{{Key:"created_at",Value: -1}})

// 		cursor, err := cartCollection.Find(ctx, bson.M{}, findOptions)
// 		if err != nil{
// 			log.Printf("Error fetching cart: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching cart data"})
// 			return
// 		}
// 		defer cursor.Close(ctx)

// 		if err := cursor.All(ctx, &carts); err != nil{
// 			log.Printf("Error decoding products: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error when decode data"})
// 			return
// 		}

// 		for i := range carts {
//             var items []models.CartItem

// 			itemCursor, err := cartCollection.Aggregate(ctx, mongo.Pipeline{
// 				bson.D{{Key: "$match", Value: bson.M{"_id": carts[i].ID}}},
// 				bson.D{{Key: "$unwind", Value: "$items"}},
// 				bson.D{{Key: "$project", Value: bson.M{
// 					"product_id": "$items.product_id",
// 					"quantity":   "$items.quantity",
// 					"price":      "$items.price",
// 					"name":       "$items.name",
// 					"image_url":  "$items.image_url",
// 				}}},
// 			})
//             if err != nil {
//                 log.Printf("Error fetching cart items: %v", err)
//                 continue
//             }
//             defer itemCursor.Close(ctx)

//             if err := itemCursor.All(ctx, &items); err != nil {
//                 log.Printf("Error decoding cart items: %v", err)
//                 continue
//             }

//             carts[i].Items = items
//         }

// 		pages := int(math.Ceil(float64(total) / float64(limit)))

// 		response := gin.H{
//             "data":     carts,
//             "total":    total,
//             "page":     page,
//             "pages":    pages,
//             "has_next": page < pages,
//             "has_prev": page > 1,
//         }

// 		log.Printf("Sending response with %d carts", len(carts))
//         c.JSON(http.StatusOK, response)
// 	}
// }

// // func  GetCart() gin.HandlerFunc{
// // 	return func(c *gin.Context){
// // 		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
// // 		defer cancel()

// // 		userID := c.GetHeader("user_id")
// // 		if userID == ""{
// // 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get userID"})
// // 			return
// // 		}

// // 		CheckUserRole(c)
// // 		if c.IsAborted(){
// // 			return
// // 		}

// // 		productID := c.Param("id")
// // 		if productID == ""{
// // 			c.JSON(http.StatusBadRequest, gin.H{"error": "Product id not found"})
// // 			return
// // 		}

// // 		log.Printf(productID)

// // 		productReq := &pb.ProductRequest{Id: productID}
// // 		productRes, err := productClient.GetProductInfo(ctx, productReq)
// // 		log.Printf(productReq.String())
// // 		log.Printf(productRes.String())
// // 		if err != nil{
// // 			log.Printf(err.Error())
// // 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product data"})
// // 			return
// // 		}

// // 		id, err := primitive.ObjectIDFromHex(productID)
// // 		if err != nil{
// // 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Product info"})
// // 			return
// // 		}

// // 		item := models.CartItem{
// // 			ProductID:  id,
// // 			Name: productRes.Name,
// // 			Price: float64(productRes.GetPrice()),
// // 			Quantity: int(productRes.GetQuantity()),
// // 			ImageUrl: productRes.ImageUrl,
// // 			Description: productRes.Description,
// // 		}

// // 		c.JSON(http.StatusOK, gin.H{
// // 			"user_id": userID,
// // 			"products" : item,
// // 		})
// // 	}
// // }

// func GetCart() gin.HandlerFunc{
// 	return func(c *gin.Context){
// 		context, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
// 		defer cancel()

// 		CheckUserRole(c)
// 		if c.IsAborted(){
// 			return
// 		}

// 		uid := c.GetHeader("user_id")
// 		if uid == ""{
// 			log.Printf("User id not found")
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "User id not found"})
// 			return
// 		}

// 		userId, err := primitive.ObjectIDFromHex(uid)
// 		if err != nil{
// 			log.Printf("Invalid user id: %v", err)
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
// 			return
// 		}

// 		var carts models.Cart

// 		filter := bson.M{"user_id": userId}
// 		err = cartCollection.FindOne(context, filter).Decode(&carts)

// 		if err != nil{
// 			log.Printf("Failed to find cart: %v", err)
// 			if err == mongo.ErrNoDocuments{
// 				c.JSON(http.StatusOK, gin.H{"message": "Cart is empty"})
// 				return
// 			}
// 			log.Printf("Error fetching cart: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
// 			return
// 		}
// 		log.Printf("Sending cart data for user %s", uid)
// 		c.JSON(http.StatusOK, gin.H{
// 			"user_id": uid,
// 			"products": carts.Items,
// 		})
// 	}
// }

// func DeleteProductFromCart() gin.HandlerFunc{
// 	return func(c *gin.Context){
// 		var ctx, cancel = context.WithTimeout(c.Request.Context(), 100*time.Second)
// 		defer cancel()

// 		userID := c.GetHeader("user_id")
// 		if userID == ""{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "User id not found"})
// 			return
// 		}
// 		CheckUserRole(c)
// 		if c.IsAborted(){
// 			return
// 		}

// 		productID := c.Param("id")
// 		if productID == ""{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Product id not found"})
// 			return
// 		}

// 		objID, err := primitive.ObjectIDFromHex(productID)
// 		if err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
// 			return
// 		}

// 		userId, err := primitive.ObjectIDFromHex(userID)
// 		if err != nil{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
// 			return
// 		}
// 		filter := bson.M{
//             "user_id": userId,
//             "items": bson.M{
//                 "$elemMatch": bson.M{
//                     "product_id": objID,
//                 },
//             },
//         }

// 		update := bson.M{
//             "$pull": bson.M{
//                 "items": bson.M{
//                     "product_id": objID,
//                 },
//             },
//         }

// 		result, err := cartCollection.UpdateOne(ctx, filter, update)
// 		if err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
// 			return
// 		}

// 		if result.ModifiedCount == 0{
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Delete product successfully"})
// 	}
// }

// func (c *CartController) ClearCart(ctx context.Context, userID primitive.ObjectID) error {
// 	// Implement the logic to clear the cart for the given userID
// 	filter := bson.M{"user_id": userID}
// 	update := bson.M{"$set": bson.M{"items": []models.CartItem{}}}

// 	_, err := cartCollection.UpdateOne(ctx, filter, update)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

package controller

import (
	"context"
	"net/http"
	"strconv"

	logger "cart-service/log"
	"cart-service/service"

	"github.com/gin-gonic/gin"
)

type CartController struct {
	cartService service.CartService
}

func NewCartController(cartService service.CartService) *CartController {
	return &CartController{
		cartService: cartService,
	}
}

// func CheckUserRole(c *gin.Context) {
// 	userRole := c.GetHeader("user_type")
// 	if userRole != "USER" {
// 		logger.Err("Unauthorized access attempt", nil)
// 		c.JSON(http.StatusUnauthorized, gin.H{"error" : "You don't have permission"})
// 		c.Abort()
// 		return
// 	}
// }

// func CheckSellerRole(c *gin.Context) {
// 	sellerRole := c.GetHeader("user_type")
// 	if sellerRole != "SELLER" {
// 		logger.Err("Unauthorized access attempt", nil)
// 		c.JSON(http.StatusUnauthorized, gin.H{"error" : "You don't have permission"})
// 		c.Abort()
// 		return
// 	}
// }

func (ctrl *CartController) AddToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// CheckUserRole(c)
		if c.IsAborted() {
			return
		}

		productID := c.Param("id")
		if productID == "" {
			logger.Err("Product id not found", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product id not found"})
			return
		}

		var requestBody struct {
			Quantity int `json:"quantity" binding:"required"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			logger.Err("Failed to bind JSON", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get quantity"})
			return
		}

		uid := c.GetHeader("user_id")
		if uid == "" {
			logger.Err("User id not found", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "User id not found"})
			return
		}

		err := ctrl.cartService.AddToCart(c, uid, productID, requestBody.Quantity)
		if err != nil {
			logger.Err("Failed to add product to cart", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product to cart"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Item added to cart successfully"})
		logger.Info("Item added to cart successfully", logger.Str("product_id", productID), logger.Str("user_id", uid))
	}
}

func (ctrl *CartController) GetCartSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		// CheckSellerRole(c)
		if c.IsAborted() {
			return
		}

		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil {
			logger.Err("Invalid page parameter, using default", err)
			page = 1
		}

		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil {
			logger.Err("Invalid limit parameter, using default", err)
			limit = 10
		}

		logger.Info("Pagination", logger.Int("page", page), logger.Int("limit", limit))

		carts, total, pages, hasNext, hasPrev, err := ctrl.cartService.GetAllCarts(c, page, limit)
		if err != nil {
			logger.Err("Error fetching carts", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch carts"})
			return
		}

		response := gin.H{
			"data":     carts,
			"total":    total,
			"page":     page,
			"pages":    pages,
			"has_next": hasNext,
			"has_prev": hasPrev,
		}

		c.JSON(http.StatusOK, response)
		logger.Info("Sending response with carts", logger.Int("cart_count", len(carts)))
	}
}

func (ctrl *CartController) GetCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// CheckUserRole(c)
		if c.IsAborted() {
			return
		}

		uid := c.GetHeader("user_id")
		if uid == "" {
			logger.Err("User id not found", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "User id not found"})
			return
		}

		cart, err := ctrl.cartService.GetUserCart(c, uid)
		if err != nil {
			if err.Error() == "mmongo: no documents in result" {
				logger.Info("Cart is empty", logger.Str("user_id", uid))
				c.JSON(http.StatusOK, gin.H{"message": "Cart is empty"})
				return
			}
			logger.Err("Error fetching cart", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
			return
		}

		logger.Info("Sending cart data", logger.Str("user_id", uid), logger.Int("cart_item_count", len(cart.Items)))
		c.JSON(http.StatusOK, gin.H{
			"user_id":  uid,
			"products": cart.Items,
		})
	}
}

func (ctrl *CartController) DeleteProductFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {

		userID := c.GetHeader("user_id")
		if userID == "" {
			logger.Err("User id not found", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "User id not found"})
			return
		}

		// CheckUserRole(c)
		if c.IsAborted() {
			return
		}

		productID := c.Param("id")
		if productID == "" {
			logger.Err("Product id not found", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product id not found"})
			return
		}

		err := ctrl.cartService.DeleteProductFromCart(c, userID, productID)
		if err != nil {
			if err.Error() == "product not found in cart" {
				logger.Info("Product not found in cart", logger.Str("user_id", userID), logger.Str("product_id", productID))
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found in cart"})
			} else {
				logger.Err("Failed to delete product from cart", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product from cart"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product deleted from cart successfully"})
		logger.Info("Product deleted from cart successfully", logger.Str("user_id", userID), logger.Str("product_id", productID))
	}
}

func (ctrl *CartController) ClearCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetHeader("user_id")
		if uid == "" {
			logger.Err("User id not found", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "User id not found"})
			return
		}

		// CheckUserRole(c)
		if c.IsAborted() {
			return
		}

		err := ctrl.cartService.ClearCart(c, uid)
		if err != nil {
			logger.Err("Error clearing cart", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cart cleared successfully"})
		logger.Info("Cart cleared successfully", logger.Str("user_id", uid))
	}
}

func (ctrl *CartController) InternalClearCart(userID string) error {
	return ctrl.cartService.ClearCart(context.Background(), userID)
}
