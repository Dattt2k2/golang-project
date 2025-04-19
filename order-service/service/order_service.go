package service

import (
    "context"
    "log"
    "time"

    cartpb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
    productpb "github.com/Dattt2k2/golang-project/module/gRPC-Product/service"
    "github.com/Dattt2k2/golang-project/order-service/kafka"
    "github.com/Dattt2k2/golang-project/order-service/models"
    "github.com/Dattt2k2/golang-project/order-service/repositories"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderService struct {
    orderRepo   *repositories.OrderRepository
}


func NewOrderService(orderRepo *repositories.OrderRepository) *OrderService {
    return &OrderService{
        orderRepo:   orderRepo,
    }
}

func (s *OrderService) CreateOrderFromCart(ctx context.Context, userID primitive.ObjectID, paymentMethod, shippingAddress string) (*models.Order, error) {
    // Get cart items using gRPC
    cartClient := CartServiceConnection()
    if cartClient == nil {
        return nil, ErrCartServiceUnavailable
    }
    
    req := &cartpb.CartRequest{
        UserId: userID.Hex(),
    }
    
    resp, err := cartClient.GetCartItems(ctx, req)
    if err != nil {
        log.Printf("Failed to get cart items: %v", err)
        return nil, err
    }
    
    productClient := ProductServiceConnection()
    if productClient == nil {
        return nil, ErrProductServiceUnavailable
    }

    // Convert cart items to order items
    var orderItems []models.OrderItem
    var totalPrice float64 = 0
    
    for _, item := range resp.Items {
        productID, err := primitive.ObjectIDFromHex(item.ProductId)
        if err != nil {
            return nil, err
        }

        stockReq := &productpb.ProductRequest{
            Id: productID.Hex(),
        }

        stockResp, err  := productClient.CheckStock(ctx, stockReq)
        if err != nil{
            return nil, NewServiceError("Failed to check stock")
        }

        if !stockResp.InStock || item.Quantity > int32(stockResp.AvailableQuantity){
            return nil, NewServiceError("Product is out of stock")
        }
        
        orderItem := models.OrderItem{
            ProductID: productID,
            Name:      item.Name,
            Quantity:  int(item.Quantity),
            Price:     float64(item.Price),
        }
        
        orderItems = append(orderItems, orderItem)
        // totalPrice += float64(item.Quantity) * float64(item.Price)
        totalPrice = calculateTotalPrice(orderItems)
    }
    
    // Create new order
    now := time.Now()
    newOrder := models.Order{
        ID:            primitive.NewObjectID(),
        UserID:        userID,
        Items:         orderItems,
        TotalPrice:    totalPrice,
        Status:        "PENDING",
        Source:        "CART",
        PaymentMethod: "COD", // Default
        PaymentStatus: "PENDING",
        Created_at:    now,
        Updated_at:    now,
    }
    
    // Save order to database
    _, err = s.orderRepo.CreateOrder(ctx, newOrder)
    if err != nil {
        return nil, err
    }
    
    // Publish Kafka event
    if err := kafka.ProduceOrderSuccessEvent(ctx, newOrder); err != nil {
        log.Printf("Warning: Failed to produce order created event: %v", err)
        // Continue as the order was successfully created
    }
    
    return &newOrder, nil
}

type OrderDirectRequest struct {
    UserID          string              `json:"user_id"`
    Items           []OrderItemRequest  `json:"items"`
    Source          string              `json:"source"`
    PaymentMethod   string              `json:"payment_method"`
    ShippingAddress string              `json:"shipping_address"`
}

type OrderItemRequest struct {
    ProductID string  `json:"product_id"`
    Name      string  `json:"name"`
    Quantity  int     `json:"quantity"`
    Price     float64 `json:"price"`
}


// CreateOrderDirect creates an order directly from the provided request
func (s *OrderService) CreateOrderDirect(ctx context.Context, req OrderDirectRequest) (*models.Order, error) {
    // Convert user ID
    userID, err := primitive.ObjectIDFromHex(req.UserID)
    if err != nil {
        return nil, err
    }
    
    productClient := ProductServiceConnection()
    if productClient == nil {
        return nil, NewServiceError("Product service unavailable")
    }

    // Convert items
    var orderItems []models.OrderItem
    var totalPrice float64 = 0
    
    for _, item := range req.Items {
        productID, err := primitive.ObjectIDFromHex(item.ProductID)
        if err != nil {
            return nil, err
        }

        stockReq := &productpb.ProductRequest{
            Id: productID.Hex(),
        }

        stockResp, err := productClient.CheckStock(ctx, stockReq)
        if err != nil{
            return nil, NewServiceError("Failed to check stock")
        }

        if !stockResp.InStock || item.Quantity > int(stockResp.AvailableQuantity){
            return nil, NewServiceError("Product is out of stock")
        }

        orderItem := models.OrderItem{
            ProductID: productID,
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
    }
    
    // Create new order
    now := time.Now()
    newOrder := models.Order{
        ID:              primitive.NewObjectID(),
        UserID:          userID,
        Items:           orderItems,
        TotalPrice:      totalPrice,
        Status:          initialStatus,
        PaymentMethod:   req.PaymentMethod,
        PaymentStatus:   paymentStatus,
        ShippingAddress: req.ShippingAddress,
        Source:          req.Source,
        Created_at:      now,
        Updated_at:      now,
    }
    
    // Save order to database
    _, err = s.orderRepo.CreateOrder(ctx, newOrder)
    if err != nil {
        return nil, err
    }
    
    // Publish Kafka event
    if err := kafka.ProduceOrderSuccessEvent(ctx, newOrder); err != nil {
        log.Printf("Warning: Failed to produce order created event: %v", err)
        // Continue as the order was successfully created
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
            log.Printf("Warning: Failed to get items for order %s: %v", orders[i].ID.Hex(), err)
            continue
        }
        orders[i].Items = items
    }
    
    return orders, total, pages, hasNext, hasPrev, nil
}


// GetUserOrders retrieves orders for a specific user with pagination
func (s *OrderService) GetUserOrders(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]models.Order,int64, int, bool, bool, error){
    orders, total, err := s.orderRepo.FindOrdersByUserID(ctx, userID, page, limit)
    if err != nil{
        return nil, 0, 0, false, false, err
    }

    pages := calculatePages(total, int64(limit))
    hasNext := page < pages
    hasPrev := page > 1

    for i := range orders {
        items, err := s.orderRepo.GetOrderItems(ctx, orders[i].ID)
        if err != nil{
            log.Printf("Warning: Failed to get items for order %s: %v", orders[i].ID.Hex(), err)
            continue
        }
        orders[i].Items = items 
    }

    return orders, total, pages, hasNext, hasPrev, nil
}


func calculateTotalPrice(items []models.OrderItem) float64 {
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

// Add error definitions
var (
    ErrCartServiceUnavailable = NewServiceError("Cart service unavailable")
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