package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"order-service/kafka"
	logger "order-service/log"
	"order-service/models"
	"order-service/repositories"

	productpb "module/gRPC-Product/service"
	cartpb "module/gRPC-cart/service"

	"gorm.io/datatypes"
)

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	VendorID  string  `json:"vendor_id"`
}
type OrderService struct {
	orderRepo *repositories.OrderRepository
}

func NewOrderService(orderRepo *repositories.OrderRepository) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
	}
}

func (s *OrderService) CreateOrderFromCart(ctx context.Context, userID string, source, paymentMethod, shippingAddress string, selectedProductIDs []string) (*models.Order, error) {
	// Get cart items using gRPC
	grpcClients := GetGRPCClients()

	cartClient := grpcClients.CartClient
	if cartClient == nil {
		return nil, ErrCartServiceUnavailable
	}

	req := &cartpb.CartRequest{
		UserId: userID,
	}

	resp, err := cartClient.GetCartItems(ctx, req)
	if err != nil {
		logger.Err("Failed to get cart items", err)
		return nil, err
	}

	productClient := grpcClients.ProductClient
	if productClient == nil {
		return nil, ErrProductServiceUnavailable
	}

	filteredItems := resp.Items
	if len(selectedProductIDs) > 0 {
		idSet := make(map[string]struct{})
		for _, id := range selectedProductIDs {
			idSet[id] = struct{}{}
		}
		var temp []*cartpb.CartItem
		for _, item := range resp.Items {
			if _, ok := idSet[item.ProductId]; ok {
				temp = append(temp, item)
			}
		}
		filteredItems = temp
	}

	// Convert cart items to order items
	var orderItems []OrderItem
	var totalPrice float64 = 0

	for _, item := range filteredItems {

		stockReq := &productpb.ProductRequest{
			Id: item.ProductId,
		}

		stockResp, err := productClient.CheckStock(ctx, stockReq)
		if err != nil {
			return nil, NewServiceError("Failed to check stock")
		}

		if !stockResp.InStock || item.Quantity > int32(stockResp.AvailableQuantity) {
			return nil, NewServiceError("Product is out of stock")
		}

		vendorID := item.VendorId
		if vendorID == "" {
			productReq := &productpb.ProductRequest{Id: item.ProductId}
			productResp, err := productClient.GetBasicInfo(ctx, productReq)
			if err == nil && productResp.VendorId != "" {
				vendorID = productResp.VendorId
			}
		}

		orderItem := OrderItem{
			VendorID:  vendorID,
			ProductID: item.ProductId,
			Name:      item.Name,
			Quantity:  int(item.Quantity),
			Price:     float64(item.Price),
		}

		orderItems = append(orderItems, orderItem)
		// totalPrice += float64(item.Quantity) * float64(item.Price)
		totalPrice = calculateTotalPrice(orderItems)
	}

	itemsJSON, err := json.Marshal(orderItems)
	if err != nil {
		return nil, err
	}

	initialStatus := "PENDING"
	paymentStatus := "PENDING"

	if paymentMethod == "stripe" {
		initialStatus = "AWAITING_FOR_PAYMENT"
		paymentStatus = "PENDING"
	} else if paymentMethod == "COD" {
		initialStatus = "CONFIRMED"
		paymentStatus = "COD_PENDING"
	}
	newOrder := models.Order{
		UserID:          userID,
		Items:           datatypes.JSON(itemsJSON),
		TotalPrice:      totalPrice,
		Status:          initialStatus,
		Source:          source,
		PaymentMethod:   paymentMethod,
		PaymentStatus:   paymentStatus,
		ShippingAddress: shippingAddress,
	}

	// Save order to database
	createdOrder, err := s.orderRepo.CreateOrder(ctx, newOrder)
	if err != nil {
		return nil, err
	}

	if paymentMethod == "stripe" {
		err = s.requestPayment(ctx, createdOrder, orderItems)
		if err != nil {
			s.orderRepo.UpdateOrderStatus(ctx, createdOrder.ID, "PAYMENT_FAILED")
			return nil, NewServiceError("Failed to initiate payment")
		}
	} else if paymentMethod == "COD" {
		if err := kafka.ProduceOrderSuccessEvent(ctx, *createdOrder); err != nil {
			logger.Err("Failed to produce order created event", err)
		}
	}

	return createdOrder, nil
}

func determinePrimaryVendor(orderItems []OrderItem) string {
	vendorAmounts := make(map[string]float64)

	for _, item := range orderItems {
		if item.VendorID != "" {
			vendorAmounts[item.VendorID] += item.Price * float64(item.Quantity)
		}
	}

	var primaryVendor string
	var maxAmount float64

	for vendorID, amount := range vendorAmounts {
		if amount > maxAmount {
			maxAmount = amount
			primaryVendor = vendorID
		}
	}

	return primaryVendor
}

// Calculate vendor breakdown for multi-vendor orders
func calculateVendorBreakdownWithFee(orderItems []OrderItem, platformFeeRate float64) map[string]map[string]float64 {
	vendorBreakdown := make(map[string]map[string]float64)

	for _, item := range orderItems {
		if item.VendorID != "" {
			itemTotal := item.Price * float64(item.Quantity)
			platformFee := itemTotal * platformFeeRate
			vendorAmount := itemTotal - platformFee

			if _, exists := vendorBreakdown[item.VendorID]; !exists {
				vendorBreakdown[item.VendorID] = make(map[string]float64)
			}

			vendorBreakdown[item.VendorID]["total_amount"] += itemTotal
			vendorBreakdown[item.VendorID]["platform_fee"] += platformFee
			vendorBreakdown[item.VendorID]["vendor_amount"] += vendorAmount
		}
	}

	return vendorBreakdown
}

func (s *OrderService) requestPayment(ctx context.Context, order *models.Order, orderItems []OrderItem) error {
	platformFeeRate := 0.05
	platformFee := order.TotalPrice * platformFeeRate
	vendorAmount := order.TotalPrice - platformFee

	// Get detailed vendor breakdown
	vendorBreakdownWithFee := calculateVendorBreakdownWithFee(orderItems, platformFeeRate)
	vendorBreakdownJSON, _ := json.Marshal(vendorBreakdownWithFee)

	// Determine primary vendor for Stripe Connect (vendor with highest amount)
	primaryVendor := determinePrimaryVendor(orderItems)

	vendorStripeAccountID := ""
	if primaryVendor != "" {
		// TODO: Call payment-service API to get vendor's Stripe account ID
		// For now, use placeholder format until vendor registration is complete
		vendorStripeAccountID = "acct_" + primaryVendor
	}

	paymentReq := kafka.PaymentRequestEvent{
		OrderID:               strconv.FormatUint(uint64(order.ID), 10),
		UserID:                order.UserID,
		Amount:                order.TotalPrice,
		PaymentMethod:         order.PaymentMethod,
		Currency:              "vnd",
		Description:           "Payment for order #" + strconv.FormatUint(uint64(order.ID), 10),
		VendorID:              primaryVendor,
		VendorStripeAccountID: vendorStripeAccountID,
		VendorAmount:          vendorAmount,
		PlatformFee:           platformFee,
		VendorBreakdown:       string(vendorBreakdownJSON),
	}

	return kafka.ProducePaymentRequestEvent(ctx, paymentReq)
}

func (s *OrderService) ReleasePaymentToVendor(ctx context.Context, orderID uint) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	// Only release if order is delivered and payment is held
	if order.Status != "DELIVERED" || order.PaymentStatus != "HELD" {
		logger.Logger.Warnf("Cannot release payment for order %d: status=%s, payment_status=%s",
			orderID, order.Status, order.PaymentStatus)
		return nil
	}

	// Parse order items to get vendor breakdown
	var items []OrderItem
	if err := json.Unmarshal(order.Items, &items); err != nil {
		logger.Err("Failed to unmarshal order items", err)
		return err
	}

	// Calculate vendor breakdown with fees
	platformFeeRate := 0.05
	vendorBreakdown := calculateVendorBreakdownWithFee(items, platformFeeRate)

	// Capture the held payment in Stripe
	if order.PaymentMethod == "stripe" && order.PaymentIntentID != nil {
		captureEvent := kafka.PaymentCaptureEvent{
			OrderID:   strconv.FormatUint(uint64(orderID), 10),
			PaymentID: *order.PaymentIntentID,
			Amount:    order.VendorAmount,
			Timestamp: time.Now().Unix(),
		}

		if err := kafka.ProducePaymentCaptureEvent(ctx, captureEvent); err != nil {
			logger.Err("Failed to produce payment capture event", err)
			return err
		}
	}

	// Update order status
	releaseTime := time.Now()
	updates := map[string]interface{}{
		"payment_status":       "RELEASED",
		"status":               "PAYMENT_RELEASED",
		"payment_release_date": releaseTime,
		"updated_at":           time.Now(),
	}

	if err := s.orderRepo.UpdateOrderFields(ctx, orderID, updates); err != nil {
		return err
	}

	// Send payment release event for EACH vendor
	for vendorID, amounts := range vendorBreakdown {
		vendorPaymentEvent := kafka.VendorPaymentEvent{
			OrderID:     strconv.FormatUint(uint64(orderID), 10),
			VendorID:    vendorID,
			Amount:      amounts["vendor_amount"], // Amount after platform fee
			PlatformFee: amounts["platform_fee"],
			ReleaseDate: releaseTime.Unix(),
			Timestamp:   time.Now().Unix(),
		}

		if err := kafka.ProduceVendorPaymentEvent(ctx, vendorPaymentEvent); err != nil {
			logger.Err("Failed to produce vendor payment event for vendor "+vendorID, err)
			// Continue với vendors khác nếu 1 vendor fail
		} else {
			logger.Logger.Infof("Payment released to vendor %s for order %d, amount: %.2f",
				vendorID, orderID, amounts["vendor_amount"])
		}
	}

	logger.Logger.Infof("Payment released to all vendors for order %d", orderID)
	return nil
}

// Helper function to get vendors from order items
func (s *OrderService) getVendorsFromOrder(orderID uint) ([]string, error) {
	order, err := s.orderRepo.GetOrderByID(context.Background(), orderID)
	if err != nil {
		return nil, err
	}

	var items []OrderItem
	if err := json.Unmarshal(order.Items, &items); err != nil {
		return nil, err
	}

	vendorSet := make(map[string]bool)
	for _, item := range items {
		if item.VendorID != "" {
			vendorSet[item.VendorID] = true
		}
	}

	var vendors []string
	for vendorID := range vendorSet {
		vendors = append(vendors, vendorID)
	}

	return vendors, nil
}

// Check if vendor owns any items in the order
func (s *OrderService) isVendorInOrder(orderID uint, vendorID string) (bool, error) {
	vendors, err := s.getVendorsFromOrder(orderID)
	if err != nil {
		return false, err
	}

	for _, v := range vendors {
		if v == vendorID {
			return true, nil
		}
	}

	return false, nil
}

func (s *OrderService) CapturePayment(ctx context.Context, orderID uint, paymentID string) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	captureEvent := kafka.PaymentCaptureEvent{
		OrderID:   strconv.FormatUint(uint64(orderID), 10),
		PaymentID: paymentID,
		Amount:    order.TotalPrice,
		Timestamp: time.Now().Unix(),
	}

	return kafka.ProducePaymentCaptureEvent(ctx, captureEvent)
}

func (s *OrderService) CancelPayment(ctx context.Context, orderID uint, paymentID, reason string) error {
	cancelEvent := kafka.PaymentCancelEvent{
		OrderID:   strconv.FormatUint(uint64(orderID), 10),
		PaymentID: paymentID,
		Reason:    reason,
		Timestamp: time.Now().Unix(),
	}

	return kafka.ProducePaymentCancelEvent(ctx, cancelEvent)
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID uint, userID string, role string) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return NewServiceError("Failed to get order")
	}

	if role == "USER" && order.UserID != userID {
		return NewServiceError("You are not authorized to cancel this order")
	}

	if order.Status == "CANCELED" {
		return NewServiceError("Order already canceled")
	}

	if order.Status == "DELIVERED" || order.Status == "DELIVERING" {
		return NewServiceError("Order already delivered or delivering")
	}

	if order.Status == "CONFIRMED" {
		if err := kafka.ProduceOrderReturnedEvent(ctx, *order); err != nil {
			logger.Err("Failed to produce order returned event", err)
		}
	}

	if order.PaymentMethod == "stripe" && order.PaymentIntentID != nil {
		if err := s.CancelPayment(ctx, order.ID, "", "Order canceled by user"); err != nil {
			logger.Err("Failed to send payment cancel event", err)
			// Continue with order cancellation even if payment cancel fails
		}
	}

	err = s.orderRepo.UpdateOrderStatus(ctx, orderID, "CANCELED")
	if err != nil {
		return err
	}

	order, _ = s.orderRepo.GetOrderByID(ctx, orderID)
	_ = kafka.ProduceOrderReturnedEvent(ctx, *order)
	return nil
}

type OrderDirectRequest struct {
	UserID          string             `json:"user_id"`
	Items           []OrderItemRequest `json:"items"`
	Source          string             `json:"source"`
	PaymentMethod   string             `json:"payment_method"`
	ShippingAddress string             `json:"shipping_address"`
}

type OrderItemRequest struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// CreateOrderDirect creates an order directly from the provided request
func (s *OrderService) CreateOrderDirect(ctx context.Context, req OrderDirectRequest) (*models.Order, error) {

	productClient := ProductServiceConnection()
	if productClient == nil {
		return nil, NewServiceError("Product service unavailable")
	}

	// Convert items
	var orderItems []OrderItem
	var totalPrice float64 = 0

	for _, item := range req.Items {

		stockReq := &productpb.ProductRequest{
			Id: item.ProductID,
		}

		stockResp, err := productClient.CheckStock(ctx, stockReq)
		if err != nil {
			return nil, NewServiceError("Failed to check stock")
		}

		if !stockResp.InStock || item.Quantity > int(stockResp.AvailableQuantity) {
			return nil, NewServiceError("Product is out of stock")
		}

		vendorID := ""
		productReq := &productpb.ProductRequest{Id: item.ProductID}
		productResp, err := productClient.GetBasicInfo(ctx, productReq)
		if err == nil && productResp.VendorId != "" {
			vendorID = productResp.VendorId
		}
		orderItem := OrderItem{
			VendorID:  vendorID,
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}

		orderItems = append(orderItems, orderItem)
		// totalPrice += float64(item.Quantity) * item.Price
		totalPrice = calculateTotalPrice(orderItems)
	}

	// Set payment details and status
	initialStatus := "PENDING"
	paymentStatus := "PENDING"

	if req.PaymentMethod == "COD" {
		initialStatus = "PROCESSING"
		paymentStatus = "PENDING_VERIFICATION"
	} else if req.PaymentMethod == "stripe" {
		initialStatus = "AWAITING_FOR_PAYMENT"
		paymentStatus = "PENDING"
	}

	itemsJSON, err := json.Marshal(orderItems)
	if err != nil {
		return nil, err
	}

	newOrder := models.Order{
		UserID:          req.UserID,
		Items:           datatypes.JSON(itemsJSON),
		TotalPrice:      totalPrice,
		Status:          initialStatus,
		PaymentMethod:   req.PaymentMethod,
		PaymentStatus:   paymentStatus,
		ShippingAddress: req.ShippingAddress,
		Source:          req.Source,
	}

	// Save order to database
	createdOrder, err := s.orderRepo.CreateOrder(ctx, newOrder)
	if err != nil {
		return nil, err
	}

	if req.PaymentMethod == "stripe" {
		err = s.requestPayment(ctx, createdOrder, orderItems)
		if err != nil {
			s.orderRepo.UpdateOrderStatus(ctx, createdOrder.ID, "PAYMENT_FAILED")
			return nil, NewServiceError("Failed to initiate payment")
		}
	} else if req.PaymentMethod == "COD" {
		if err := kafka.ProduceOrderSuccessEvent(ctx, *createdOrder); err != nil {
			logger.Err("Failed to produce order created event", err)
		}
	}

	return &newOrder, nil
}

// AdminGetOrders retrieves all orders with pagination
func (s *OrderService) AdminGetOrders(ctx context.Context, page, limit int) ([]models.Order, int64, int, bool, bool, error) {
	orders, total, err := s.orderRepo.FindOrders(ctx, page, limit)
	if err != nil {
		return nil, 0, 0, false, false, err
	}

	// Calculate pagination info
	pages := calculatePages(total, int64(limit))
	hasNext := page < pages
	hasPrev := page > 1

	// Fetch items for each order
	for i := range orders {
		items, err := s.orderRepo.GetOrderItems(ctx, orders[i].ID)
		if err != nil {
			continue
		}
		orders[i].Items = items
	}

	return orders, total, pages, hasNext, hasPrev, nil
}

// GetUserOrders retrieves orders for a specific user with pagination
func (s *OrderService) GetUserOrders(ctx context.Context, userID string, page, limit int) ([]models.Order, int64, int, bool, bool, error) {
	orders, total, err := s.orderRepo.FindOrdersByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, 0, false, false, err
	}

	pages := calculatePages(total, int64(limit))
	hasNext := page < pages
	hasPrev := page > 1

	for i := range orders {
		items, err := s.orderRepo.GetOrderItems(ctx, orders[i].ID)
		if err != nil {
			continue
		}
		orders[i].Items = items
	}

	return orders, total, pages, hasNext, hasPrev, nil
}

func (s *OrderService) HandlePaymentSuccess(ctx context.Context, orderID uint, paymentIntentID string) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	// Calculate platform fee and vendor amount (example: 5% platform fee)
	platformFeeRate := 0.05
	platformFee := order.TotalPrice * platformFeeRate
	vendorAmount := order.TotalPrice - platformFee

	// Update order with payment info - HELD status, not released yet
	updates := map[string]interface{}{
		"payment_status":    "HELD",
		"payment_intent_id": paymentIntentID,
		"status":            "PAYMENT_HELD",
		"platform_fee":      platformFee,
		"vendor_amount":     vendorAmount,
		"updated_at":        time.Now(),
	}

	if err := s.orderRepo.UpdateOrderFields(ctx, orderID, updates); err != nil {
		return err
	}

	// Reduce inventory (same as before)
	if err := kafka.ProduceOrderSuccessEvent(ctx, *order); err != nil {
		logger.Err("Failed to produce order success event after payment", err)
		return NewServiceError("Payment successful but failed to update inventory")
	}

	logger.Logger.Infof("Order %d payment held in escrow, vendor amount: %.2f", orderID, vendorAmount)
	return nil
}

func (s *OrderService) HandlePaymentFailure(ctx context.Context, orderID uint, reason string) error {
	if err := s.orderRepo.UpdatePaymentStatus(ctx, orderID, "FAILED"); err != nil {
		return err
	}

	if err := s.orderRepo.UpdateOrderPaymentStatus(ctx, orderID, "PAYMENT_FAILED", nil); err != nil {
		return err
	}

	logger.Logger.Infof("Order %d payment failed. Reason: %s", orderID, reason)
	return nil
}

func (s *OrderService) ConfirmDelivery(ctx context.Context, orderID uint, userID string) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	// Only order owner can confirm delivery
	if order.UserID != userID {
		return NewServiceError("Unauthorized to confirm delivery")
	}

	// Order must be shipped to confirm delivery
	if order.Status != "SHIPPED" {
		return NewServiceError("Order must be shipped before confirming delivery")
	}

	// Update order status to delivered
	deliveryTime := time.Now()
	updates := map[string]interface{}{
		"status":        "DELIVERED",
		"delivery_date": deliveryTime,
		"updated_at":    time.Now(),
	}

	if err := s.orderRepo.UpdateOrderFields(ctx, orderID, updates); err != nil {
		return err
	}

	// Trigger payment release after delivery confirmation (24h delay)
	go func() {
		time.Sleep(24 * time.Hour) // Auto-release after 24 hours
		s.ReleasePaymentToVendor(context.Background(), orderID)
	}()

	logger.Logger.Infof("Order %d delivery confirmed, payment will be released to vendor", orderID)
	return nil
}

// MarkAsShipped - Vendor marks order as shipped
func (s *OrderService) MarkAsShipped(ctx context.Context, orderID uint, vendorID string) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	// Check if vendor owns any items in this order
	isVendorInOrder, err := s.isVendorInOrder(orderID, vendorID)
	if err != nil {
		return err
	}

	if !isVendorInOrder {
		return NewServiceError("Vendor is not associated with this order")
	}

	// Check order status
	if order.Status != "PAYMENT_HELD" && order.Status != "CONFIRMED" {
		return NewServiceError("Order must be confirmed before shipping")
	}

	updates := map[string]interface{}{
		"status":     "SHIPPED",
		"updated_at": time.Now(),
	}

	if err := s.orderRepo.UpdateOrderFields(ctx, orderID, updates); err != nil {
		return err
	}

	logger.Logger.Infof("Order %d marked as shipped by vendor %s", orderID, vendorID)
	return nil
}

func calculateTotalPrice(items []OrderItem) float64 {
	var totalPrice float64
	for _, item := range items {
		totalPrice += float64(item.Quantity) * item.Price
	}
	return totalPrice
}

// Calculate the number of pages based on total items and limit
func calculatePages(total int64, limit int64) int {
	if total == 0 || limit == 0 {
		return 0
	}

	pages := int(total / limit)
	if total%limit > 0 {
		pages++
	}
	return pages
}

func (s *OrderService) GetOrderByID(ctx context.Context, orderID uint) (*models.Order, error) {
	return s.orderRepo.GetOrderByID(ctx, orderID)
}

// Add error definitions
var (
	ErrCartServiceUnavailable    = NewServiceError("Cart service unavailable")
	ErrProductServiceUnavailable = NewServiceError("Product service unavailable")
)

// ServiceError represents a service-level error
type ServiceError struct {
	message string
}

func NewServiceError(message string) *ServiceError {
	return &ServiceError{message: message}
}

func (e *ServiceError) Error() string {
	return e.message
}