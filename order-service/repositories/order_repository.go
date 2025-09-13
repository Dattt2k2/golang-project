package repositories

import (
	"context"
	"math"

	"order-service/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

// CreateOrder inserts a new order into the database
func (r *OrderRepository) CreateOrder(ctx context.Context, order models.Order) (*models.Order, error) {
	if err := r.db.WithContext(ctx).Create(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// FindOrders retrieves all orders with pagination
func (r *OrderRepository) FindOrders(ctx context.Context, page, limit int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	r.db.Model(&models.Order{}).Count(&total)
	if total == 0 {
		return []models.Order{}, 0, nil
	}

	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// FindOrdersByUserID retrieves orders for a specific user with pagination
func (r *OrderRepository) FindOrdersByUserID(ctx context.Context, userID string, page, limit int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	r.db.Model(&models.Order{}).Where("user_id = ?", userID).Count(&total)
	if total == 0 {
		return []models.Order{}, 0, nil
	}

	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// GetOrderItems retrieves items for a specific order
func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID uint) (datatypes.JSON, error) {
	var order models.Order
	err := r.db.WithContext(ctx).First(&order, orderID).Error
	if err != nil {
		return nil, err
	}
	return order.Items, nil
}

// CalculateOrderPages calculates pagination info based on total orders and limit
func CalculateOrderPages(total int64, limit int) int {
	return int(math.Ceil(float64(total) / float64(limit)))
}

// UpdateOrderStatus updates the status of an order
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID uint, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("id = ?", orderID).
		Update("status", status).Error
}

// UpdatePaymentStatus updates the payment_status field of an order
func (r *OrderRepository) UpdatePaymentStatus(ctx context.Context, orderID uint, paymentStatus string) error {
	return r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("id = ?", orderID).
		Update("payment_status", paymentStatus).Error
}

// FindOrderByID retrieves a specific order by ID
func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID uint) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).First(&order, orderID).Error
	if err != nil {
		return nil, err
	}

	return &order, nil
}
