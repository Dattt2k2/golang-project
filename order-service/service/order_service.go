package service

import (
    "context"
    "log"
    "time"

	"github.com/Dattt2k2/golang-project/order-service/repositories"
    "github.com/Dattt2k2/golang-project/order-service/kafka"
    "github.com/Dattt2k2/golang-project/order-service/models"
    cartpb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
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
    client := CartServiceConnection()
    if client == nil {
        return nil, ErrCartServiceUnavailable
    }
    
    req := &cartpb.CartRequest{
        UserId: userID.Hex(),
    }
    
    resp, err := client.GetCartItems(ctx, req)
    if err != nil {
        log.Printf("Failed to get cart items: %v", err)
        return nil, err
    }
    
    // Convert cart items to order items
    var orderItems []models.OrderItem
    var totalPrice float64 = 0
    
    for _, item := range resp.Items {
        productID, err := primitive.ObjectIDFromHex(item.ProductId)
        if err != nil {
            return nil, err
        }
        
        orderItem := models.OrderItem{
            ProductID: productID,
            Name:      item.Name,
            Quantity:  int(item.Quantity),
            Price:     float64(item.Price),
        }
        
        orderItems = append(orderItems, orderItem)
        totalPrice += float64(item.Quantity) * float64(item.Price)
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
    _, err = s.orderRepo.CreateOrder(ctx, &newOrder)
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

func (s *OrderService) CreateOrderDirect(ctx context.Context, req OrderDirectRequest) (*models.Order, error) {
    // Convert user ID
    userID, err := primitive.ObjectIDFromHex(req.UserID)
    if err != nil {
        return nil, err
    }
    
    // Convert items
    var orderItems []models.OrderItem
    var totalPrice float64 = 0
    
    for _, item := range req.Items {
        productID, err := primitive.ObjectIDFromHex(item.ProductID)
        if err != nil {
            return nil, err
        }
        
        orderItem := models.OrderItem{
            ProductID: productID,
            Name:      item.Name,
            Quantity:  item.Quantity,
            Price:     item.Price,
        }
        
        orderItems = append(orderItems, orderItem)
        totalPrice += float64(item.Quantity) * item.Price
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
    _, err = s.orderRepo.CreateOrder(ctx, &newOrder)
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

func (s *OrderService) GetOrders(ctx context.Context, page, limit int) ([]models.Order, int64, int, bool, bool, error) {
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