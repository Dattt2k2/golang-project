package controller

import (
	"context"
	"net/http"
	"strconv"
	"time"

	logger "order-service/log"
	"order-service/service"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	orderService *service.OrderService
}

func NewOrderController(orderService *service.OrderService) *OrderController {
	return &OrderController{
		orderService: orderService,
	}
}

func (ctrl *OrderController) OrderFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// CheckUserRole(c)
		if c.IsAborted() {
			return
		}

		uid := c.GetHeader("X-User-ID")

		type OrderCartRequest struct {
			Source             string   `json:"source"`
			PaymentMethod      string   `json:"paymentMethod"`
			ShippingAddress    string   `json:"shippingAddress"`
			Items 			[]struct {
				ProductId string `json:"productId"`
			}
		}

		var requestBody OrderCartRequest
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			logger.Err("Failed to bind JSON", err, logger.Str("request_body", requestBody.Source))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		if requestBody.PaymentMethod != "COD" && requestBody.PaymentMethod != "STRIPE" {
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

		var selectedProductIDs []string
		for _, item := range requestBody.Items {
			selectedProductIDs = append(selectedProductIDs, item.ProductId)
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		order, err := ctrl.orderService.CreateOrderFromCart(ctx, uid, requestBody.Source, requestBody.PaymentMethod, requestBody.ShippingAddress, selectedProductIDs)

		if err != nil {
			if err == service.ErrCartServiceUnavailable {
				logger.Err("Cart service unavailable", err, logger.Str("user_id", uid))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Cart service unavailable"})
				return
			} else {
				logger.Err("Failed to create order", err, logger.Str("user_id", uid))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":          "Order placed successfully",
			"order_id":         order.ID,
			"total_price":      order.TotalPrice,
			"payment_method":   order.PaymentMethod,
			"shipping_address": order.ShippingAddress,
			"status":           order.Status,
		})
	}
}

// Order directly from product, using product ID and quantity
// This function is used when user want to order directly from product page
func (ctrl *OrderController) OrderDirectly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// CheckUserRole(c)
		if c.IsAborted() {
			return
		}

		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			logger.Err("Failed to get userID", nil, logger.Str("user_id", userID))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
			return
		}

		var orderReq service.OrderDirectRequest
		if err := c.ShouldBindJSON(&orderReq); err != nil {
			logger.Err("Failed to bind JSON", err, logger.Str("request_body", orderReq.Source))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		orderReq.UserID = userID
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		order, err := ctrl.orderService.CreateOrderDirect(ctx, orderReq)
		if err != nil {
			logger.Err("Failed to create order", err, logger.Str("user_id", userID))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":          "Order placed successfully",
			"id":               order.ID,
			"order_id":         order.OrderID,
			"total_price":      order.TotalPrice,
			"payment_method":   order.PaymentMethod,
			"shipping_address": order.ShippingAddress,
			"status":           order.Status,
		})

	}
}
func (ctrl *OrderController) AdminGetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// CheckSellerRole(c)
		if c.IsAborted() {
			return
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			logger.Err("Failed to parse page", err, logger.Str("page", c.Query("page")))
			page = 1
		}

		limit, err := strconv.Atoi(c.Query("limit"))
		if err != nil {
			logger.Err("Failed to parse limit", err, logger.Str("limit", c.Query("limit")))
			limit = 10
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		orders, total, pages, hasNext, hasPrev, err := ctrl.orderService.AdminGetOrders(ctx, page, limit)
		if err != nil {
			logger.Err("Failed to get orders", err, logger.Str("page", strconv.Itoa(page)), logger.Str("limit", strconv.Itoa(limit)))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders"})
			return
		}

		if len(orders) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"data":     []interface{}{},
				"total":    0,
				"page":     page,
				"limit":    limit,
				"pages":    0,
				"has_next": false,
				"has_prev": false,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":     orders,
			"total":    total,
			"page":     page,
			"limit":    limit,
			"pages":    pages,
			"has_next": hasNext,
			"has_prev": hasPrev,
		})

	}
}

func (ctrl *OrderController) GetOrdersByVendor() gin.HandlerFunc {
	return func(c *gin.Context) {
		vendorID := c.GetHeader("X-User-ID")
		if vendorID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "vendor_id is required"})
			return
		}

		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
			return
		}

		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil || limit < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}

		status := c.Query("status")
		month, _ := strconv.Atoi(c.DefaultQuery("month", "0"))
		year, _ := strconv.Atoi(c.DefaultQuery("year", "0"))

		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		orders, total, totalRevenue, err := ctrl.orderService.GetOrdersByVendor(ctx, vendorID, page, limit, status, month, year)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":          orders,
			"total":         total,
			"total_revenue": totalRevenue,
			"page":          page,
			"limit":         limit,
		})
	}
}

// GetUserOrders retrieves orders for a specific user with pagination
func (ctrl *OrderController) GetUserOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// CheckUserRole(c)
		if c.IsAborted() {
			return
		}

		uid := c.GetHeader("X-User-ID")

		page := 1
		limit := 10

		pageStr := c.Query("page")
		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil {
				page = p
			}
		}

		limitStr := c.Query("limit")
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		orders, total, pages, hasNext, hasPrev, err := ctrl.orderService.GetUserOrders(ctx, uid, page, limit)
		if err != nil {
			logger.Err("Failed to get orders", err, logger.Str("user_id", uid), logger.Str("page", strconv.Itoa(page)), logger.Str("limit", strconv.Itoa(limit)))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders"})
			return
		}

		if len(orders) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"data":     []interface{}{},
				"total":    0,
				"page":     page,
				"limit":    limit,
				"pages":    0,
				"has_next": false,
				"has_prev": false,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":     orders,
			"total":    total,
			"page":     page,
			"limit":    limit,
			"pages":    pages,
			"has_next": hasNext,
			"has_prev": hasPrev,
		})

	}
}

func (ctrl *OrderController) VendorUpdateOrderStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		orderID := c.Param("id")
		vendorID := c.GetHeader("X-User-ID")

		type UpdateStatusRequest struct {
			Status string `json:"status"`
		}

		var req UpdateStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Err("Failed to bind JSON", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}


		err := ctrl.orderService.AdminUpdateOrderStatus(ctx, orderID, vendorID, req.Status)
		if err != nil {
			logger.Err("Failed to update order status", err, logger.Str("order_id", orderID), logger.Str("vendor_id", vendorID))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Order status updated successfully",
		})
	}
}

// UpdateOrderStatus - Universal endpoint for both user and vendor to update order status
func (ctrl *OrderController) UpdateOrderStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		orderID := c.Param("id")
		userID := c.GetHeader("X-User-ID")

		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		type UpdateStatusRequest struct {
			Status string `json:"status" binding:"required"`
		}

		var req UpdateStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Err("Failed to bind JSON", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}


		// Call the universal update method
		err := ctrl.orderService.UpdateOrderStatusWithPayout(ctx, orderID, userID, req.Status)
		if err != nil {
			logger.Err("Failed to update order status", err,
				logger.Str("order_id", orderID),
				logger.Str("user_id", userID))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		responseMessage := "Order status updated successfully"
		if req.Status == "DELIVERED" {
			responseMessage = "Order delivered and payment will be processed to vendor"
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  responseMessage,
			"order_id": orderID,
			"status":   req.Status,
		})
	}
}

// Cancel Order with ID
func (ctrl *OrderController) CancelOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()
		orderID := c.Param("order_id")
		userID := c.GetHeader("X-User-ID")
		err := ctrl.orderService.CancelOrder(ctx, orderID, userID)
		if err != nil {
			logger.Err("Failed to cancel order", err, logger.Str("order_id", orderID), logger.Str("user_id", userID))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Order cancelled successfully"})

	}

}

func (ctrl *OrderController) HandlePaymentSuccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get order ID from URL parameter
		orderID := c.Param("id")

		// Parse request body
		type PaymentSuccessRequest struct {
			PaymentIntentID string `json:"payment_intent_id"`
		}

		var req PaymentSuccessRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Err("Failed to bind JSON", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Call service method
		if err := ctrl.orderService.HandlePaymentSuccess(ctx, orderID, req.PaymentIntentID); err != nil {
			logger.Err("Failed to handle payment success", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment success"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Payment success processed",
			"order_id": orderID,
			"status":   "confirmed",
		})
	}
}

// HandlePaymentFailure handles payment failure callback from payment-service
func (ctrl *OrderController) HandlePaymentFailure() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get order ID from URL parameter
		orderID := c.Param("id")

		// Parse request body
		type PaymentFailureRequest struct {
			Reason string `json:"reason"`
		}

		var req PaymentFailureRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Err("Failed to bind JSON", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Call service method
		if err := ctrl.orderService.HandlePaymentFailure(ctx, orderID, req.Reason); err != nil {
			logger.Err("Failed to handle payment failure", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment failure"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Payment failure processed",
			"order_id": orderID,
			"status":   "payment_failed",
		})
	}
}

func (ctrl *OrderController) ConfirmDelivery() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")

		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := ctrl.orderService.ConfirmDelivery(ctx, orderID, userID); err != nil {
			logger.Err("Failed to confirm delivery", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Delivery confirmed successfully",
			"order_id": orderID,
			"status":   "delivered",
		})
	}
}

// MarkAsShipped - Vendor marks order as shipped
func (ctrl *OrderController) MarkAsShipped() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")

		vendorID := c.GetHeader("X-User-ID")

		if vendorID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Vendor authentication required"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := ctrl.orderService.MarkAsShipped(ctx, orderID, vendorID); err != nil {
			logger.Err("Failed to mark as shipped", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Order marked as shipped successfully",
			"order_id": orderID,
			"status":   "shipped",
		})
	}
}

func (ctrl *OrderController) GetOrderByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		order, err := ctrl.orderService.GetOrderByID(ctx, orderID)
		if err != nil {
			logger.Err("Failed to get order", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"order_id":         order.OrderID,
			"user_id":          order.UserID,
			"status":           order.Status,
			"payment_status":   order.PaymentStatus,
			"payment_method":   order.PaymentMethod,
			"total_price":      order.TotalPrice,
			"shipping_address": order.ShippingAddress,
			"created_at":       order.CreatedAt,
			"updated_at":       order.UpdatedAt,
		})
	}
}

// GetOrderStatus - Get detailed order status (for both buyer and vendor)
func (ctrl *OrderController) GetOrderStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		order, err := ctrl.orderService.GetOrderByID(ctx, orderID)
		if err != nil {
			logger.Err("Failed to get order", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order"})
			return
		}

		// Check authorization - only order owner or vendor can view
		userID := c.GetHeader("X-User-ID")
		userType := c.GetHeader("user_type")

		if userType != "ADMIN" && userType != "SELLER" {
			if order.UserID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to view this order"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"order_id":             order.ID,
			"status":               order.Status,
			"payment_status":       order.PaymentStatus,
			"payment_method":       order.PaymentMethod,
			"total_price":          order.TotalPrice,
			"platform_fee":         order.PlatformFee,
			"vendor_amount":        order.VendorAmount,
			"shipping_address":     order.ShippingAddress,
			"delivery_date":        order.DeliveryDate,
			"payment_release_date": order.PaymentReleaseDate,
			"created_at":           order.CreatedAt,
			"updated_at":           order.UpdatedAt,
		})
	}
}

// ReleasePaymentManually - Admin can manually release payment (emergency use)
func (ctrl *OrderController) ReleasePaymentManually() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")

		userType := c.GetHeader("user_type")
		if userType != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can manually release payments"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := ctrl.orderService.ReleasePaymentToVendor(ctx, orderID); err != nil {
			logger.Err("Failed to manually release payment", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Payment released to vendor successfully",
			"order_id": orderID,
		})
	}
}

func (ctrl *OrderController) GetOrderStatistics() gin.HandlerFunc {
	return func(c *gin.Context) {
		month, _ := strconv.Atoi(c.DefaultQuery("month", "0"))
		year, _ := strconv.Atoi(c.DefaultQuery("year", "0"))

		if month == 0 || year == 0 {
			now := time.Now()
			month = int(now.Month())
			year = now.Year()
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		userType := c.GetHeader("X-User-Type")
		if userType != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can access order statistics"})
			return
		}

		stats, err := ctrl.orderService.GetOrderStatistics(ctx, month, year)
		if err != nil {
			logger.Err("Failed to get order statistics", err, logger.Str("month", strconv.Itoa(month)), logger.Str("year", strconv.Itoa(year)))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order statistics"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func (ctrl *OrderController) GetShippedOrderCount() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		userType := c.GetHeader("X-User-Type")
		if userType != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can access shipped order count"})
			return
		}

		userID := c.Param("user_id")
		
		count, totalPrice , err := ctrl.orderService.GetShippedOrdersCountAndTotalPrice(ctx, userID)
		if err != nil {
			logger.Err("Failed to get shipped order count", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get shipped order count"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"shipped_order_count": count,
			"total_price": totalPrice,
		})
	}
}
