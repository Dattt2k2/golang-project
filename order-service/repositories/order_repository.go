package repositories

import (
	"context"
	"math"
	"time"

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

func (r *OrderRepository) FindOrdersByVendorID(ctx context.Context, vendorID string, page, limit int, status string, month int, year int) ([]models.Order, int64, float64, error) {
    var orders []models.Order
    var total int64
    var totalRevenue float64

    // Tạo base query
    baseQuery := r.db.WithContext(ctx).Model(&models.Order{})

    // Filter theo VendorID
    baseQuery = baseQuery.Where("items @> ?", `[{"vendor_id": "`+vendorID+`"}]`)

    // Filter theo trạng thái đơn hàng (nếu có)
    if status != "" {
        baseQuery = baseQuery.Where("status = ?", status)
    }

    // Filter theo tháng và năm (nếu có)
    if month > 0 && year > 0 {
        startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
        endDate := startDate.AddDate(0, 1, 0)
        baseQuery = baseQuery.Where("created_at >= ? AND created_at < ?", startDate, endDate)
    }

    // Đếm tổng số đơn hàng (sử dụng session riêng)
    err := baseQuery.Session(&gorm.Session{}).Count(&total).Error
    if err != nil {
        return nil, 0, 0, err
    }

    if total == 0 {
        return []models.Order{}, 0, 0, nil
    }

    // Tính tổng doanh thu (sử dụng session riêng)
    err = baseQuery.Session(&gorm.Session{}).Select("COALESCE(SUM(total_price), 0)").Scan(&totalRevenue).Error
    if err != nil {
        return nil, 0, 0, err
    }

    // Phân trang và lấy danh sách đơn hàng (sử dụng session riêng)
    offset := (page - 1) * limit
    err = baseQuery.Session(&gorm.Session{}).
        Order("created_at DESC").
        Limit(limit).
        Offset(offset).
        Find(&orders).Error
    if err != nil {
        return nil, 0, 0, err
    }

    return orders, total, totalRevenue, nil
}

func (r *OrderRepository) GetUserOrderWithProductID(ctx context.Context, userID, productID string) (models.Order, error) {
	var order models.Order

	// Query orders where user_id matches and items JSONB contains the product_id
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("items::text LIKE ?", "%"+productID+"%").
		First(&order).Error

	if err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func (r *OrderRepository) GetByOrderID(ctx context.Context, orderID string) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
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
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("order_id = ?", orderID).
		Update("status", status).Error
}

// UpdatePaymentStatus updates the payment_status field of an order
func (r *OrderRepository) UpdatePaymentStatus(ctx context.Context, orderID string, paymentStatus string) error {
	return r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("order_id = ?", orderID).
		Update("payment_status", paymentStatus).Error
}

func (r *OrderRepository) UpdateOrderPaymentStatus(ctx context.Context, orderID string, paymentStatus string, paymentIntentID *string) error {
	updates := map[string]interface{}{
		"payment_status": paymentStatus,
		"updated_at":     time.Now(),
	}

	if paymentIntentID != nil {
		updates["payment_intent_id"] = *paymentIntentID
	}

	return r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}

// FindOrderByID retrieves a specific order by ID
func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&order).Error
    if err != nil {
        return nil, err
    }
    return &order, nil
}

func (r *OrderRepository) UpdatePaymentIntentID(ctx context.Context, orderID string, paymentIntentID string) error {
	return r.db.WithContext(ctx).Model(&models.Order{}).
		Where("order_id = ?", orderID).
		Update("payment_intent_id", paymentIntentID).Error
}

func (r *OrderRepository) UpdateOrderFields(ctx context.Context, orderID string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}
