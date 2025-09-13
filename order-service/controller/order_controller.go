package controller

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"order-service/log"
	"order-service/service"
	"github.com/gin-gonic/gin"
)


type OrderController struct {
	orderService *service.OrderService
}

func NewOrderController(orderService *service.OrderService)  *OrderController{
	return &OrderController{
		orderService: orderService,
	}
}

func (ctrl *OrderController) OrderFromCart() gin.HandlerFunc{
	return func(c *gin.Context){
		// CheckUserRole(c)
		if c.IsAborted(){
			return
		}

		uid := c.GetHeader("user_id")

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

		ctx, cancel := context.WithTimeout(c.Request.Context(), 10 * time.Second)
		defer cancel()

		order, err := ctrl.orderService.CreateOrderFromCart(ctx, uid, requestBody.Source, requestBody.PaymentMethod, requestBody.ShippingAddress, requestBody.SelectedProductIDs)

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
			"order_id": order.ID,
			"total_price": order.TotalPrice,
            "payment_method": order.PaymentMethod,
            "shipping_address": order.ShippingAddress,
            "status": order.Status,
		})
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
        ctx, cancel := context.WithTimeout(c.Request.Context(), 10 * time.Second)
        defer cancel()

        order, err :=  ctrl.orderService.CreateOrderDirect(ctx, orderReq)
        if err != nil{
            logger.Err("Failed to create order", err, logger.Str("user_id", userID))
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
            return 
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Order placed successfully",
            "order_id": order.ID,
            "total_price": order.TotalPrice,
            "payment_method": order.PaymentMethod,
            "shipping_address": order.ShippingAddress,
            "status": order.Status,
        })

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

        ctx, cancel := context.WithTimeout(c.Request.Context(), 10 * time.Second)
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

        ctx, cancel := context.WithTimeout(c.Request.Context(), 10 * time.Second)
        defer cancel()

        orders, total, pages, hasNext, hasPrev, err := ctrl.orderService.GetUserOrders(ctx, uid, page, limit)
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
        ctx, cancel := context.WithTimeout(c.Request.Context(), 10 * time.Second)
        defer cancel()
        userRole := c.GetHeader("user_type")
        if userRole != "USER" && userRole != "SELLER" {
            logger.Err("Unauthorized access", nil, logger.Str("user_role", userRole))
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user type"})
            return
        }
        orderID := c.Param("order_id")
        orderIDUint, err := strconv.ParseUint(orderID, 10, 32)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
            return
        }
        userID := c.GetHeader("user_id")
        err = ctrl.orderService.CancelOrder(ctx, uint(orderIDUint), userID, userRole)
        if err != nil{
            logger.Err("Failed to cancel order", err, logger.Str("order_id", orderID), logger.Str("user_id", userID))
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return 
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Order cancelled successfully"})

    }

}


