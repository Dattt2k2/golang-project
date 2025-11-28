package repositories

import (
	"context"
	"fmt"
	"math"
	"time"

	logger "order-service/log"
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

	revenueQuery := r.db.WithContext(ctx).Model(&models.Order{})
	revenueQuery = revenueQuery.Where("items @> ?", `[{"vendor_id": "`+vendorID+`"}]`)
	revenueQuery = revenueQuery.Where("status = ?", "SHIPPED")

	if month > 0 && year > 0 {
		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, 0)
		revenueQuery = revenueQuery.Where("created_at >= ? AND created_at < ?", startDate, endDate)
	}

	err = revenueQuery.Session(&gorm.Session{}).Select("COALESCE(SUM(total_price), 0)").Scan(&totalRevenue).Error
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
	order, err := r.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.PaymentMethod == "COD" && updates["status"] == "SHIPPED" {
		updates["payment_status"] = "PAID"
	}
	return r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}

func (r *OrderRepository) UpdateOrderStatusByVendorID(ctx context.Context, orderID string, vendorID string, status string) error {
	if orderID == "" || vendorID == "" {
		return fmt.Errorf("orderID and vendorID cannot be empty")
	}

	result := r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("order_id = ? AND items @> ?", orderID, `[{"vendor_id": "`+vendorID+`"}]`).
		Update("status", status)

	if result.RowsAffected == 0 {
		logger.Err("Failed to update order status", fmt.Errorf("No rows updated: orderID=%s, vendorID=%s, status=%s", orderID, vendorID, status))
		return fmt.Errorf("no rows updated: order_id=%s, vendor_id=%s", orderID, vendorID)
	}

	return result.Error
}

func (r *OrderRepository) GetOrderStatus(ctx context.Context, orderID string) (string, string, string, error) {
	var result struct {
		Status        string
		PaymentMethod string
		PaymentStatus string
	}
	err := r.db.WithContext(ctx).
		Model(&models.Order{}).
		Select("status, payment_method, payment_status").
		Where("order_id = ?", orderID).
		Scan(&result).Error
	if err != nil {
		return "", "", "", err
	}
	return result.Status, result.PaymentMethod, result.PaymentStatus, nil

}

func (r *OrderRepository) GetOrderStatistics(ctx context.Context, month int, year int) (int64, float64, int64, float64, int64, []models.TopProduct, error) {
	var totalOrderCount int64
	var totalRevenue float64
	var prevOrderCount int64
	var prevRevenue float64
	var totalQuantity int64
	var topProducts []models.TopProduct

	buildRange := func(m, y int) (time.Time, time.Time) {
		if m <= 0 || y <= 0 {
			return time.Time{}, time.Time{}
		}
		start := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)
		return start, end
	}

	start, end := buildRange(month, year)

	q := r.db.WithContext(ctx).Model(&models.Order{}).Where("status = ?", "SHIPPED")
	if !start.IsZero() {
		q = q.Where("created_at >= ? AND created_at < ?", start, end)
	}
	if err := q.Count(&totalOrderCount).Error; err != nil {
		return 0, 0, 0, 0, 0, nil, err
	}

	revQ := r.db.WithContext(ctx).Model(&models.Order{}).Where("status = ?", "SHIPPED")
	if !start.IsZero() {
		revQ = revQ.Where("created_at >= ? AND created_at < ?", start, end)
	}
	if err := revQ.Select("COALESCE(SUM(total_price), 0)").Scan(&totalRevenue).Error; err != nil {
		return 0, 0, 0, 0, 0, nil, err
	}

	prevStart, prevEnd := buildRange(month-1, year)
	prevQ := r.db.WithContext(ctx).Model(&models.Order{}).Where("status = ?", "SHIPPED")
	if !prevStart.IsZero() {
		prevQ = prevQ.Where("created_at >= ? AND created_at < ?", prevStart, prevEnd)
	}
	if err := prevQ.Count(&prevOrderCount).Error; err != nil {
		return 0, 0, 0, 0, 0, nil, err
	}

	prevRevQ := r.db.WithContext(ctx).Model(&models.Order{}).Where("status = ?", "SHIPPED")
	if !prevStart.IsZero() {
		prevRevQ = prevRevQ.Where("created_at >= ? AND created_at < ?", prevStart, prevEnd)
	}
	if err := prevRevQ.Select("COALESCE(SUM(total_price), 0)").Scan(&prevRevenue).Error; err != nil {
		return 0, 0, 0, 0, 0, nil, err
	}

	// totalQtySQL := `
	//     SELECT COALESCE(SUM((it->>'quantity')::bigint), 0) AS total_quantity
	//     FROM orders, jsonb_array_elements(items) AS it
	//     WHERE status = 'SHIPPED'
	// `
	var args []interface{}
	sql := `
        SELECT
            it->>'product_id' AS product_id,
            it->>'name' AS name,
            SUM((it->>'quantity')::bigint) AS total_quantity,
            COALESCE(SUM(((it->>'price')::numeric) * ((it->>'quantity')::bigint)), 0) AS total_revenue,
            COUNT(DISTINCT o.id) AS total_orders
        FROM orders o, jsonb_array_elements(o.items) AS it
        WHERE o.status = 'SHIPPED'
    `
	if !start.IsZero() {
		sql += " AND o.created_at >= ? AND o.created_at < ?"
		args = append(args, start, end)
	}
	sql += `
        GROUP BY it->>'product_id', it->>'name'
        ORDER BY total_quantity DESC, total_revenue DESC
        LIMIT 5
    `

	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&topProducts).Error; err != nil {
		return 0, 0, 0, 0, 0, nil, err
	}

	return totalOrderCount, totalRevenue, prevOrderCount, prevRevenue, totalQuantity, topProducts, nil
}

func (r *OrderRepository) GetShippedOrdersCountAndTotal(ctx context.Context, userID string) (int64, float64, error) {
	var count int64
	var totalValue float64

	err := r.db.WithContext(ctx).Model(&models.Order{}).Where("status = ? AND user_id = ?", "SHIPPED", userID).Count(&count).Error
	if err != nil {
		return 0, 0, err
	}

	err = r.db.WithContext(ctx).Model(&models.Order{}).Where("status = ? AND user_id = ?", "SHIPPED", userID).Select("COALESCE(SUM(total_price), 0)").Scan(&totalValue).Error
	if err != nil {
		return 0, 0, err
	}

	return count, totalValue, nil
}
