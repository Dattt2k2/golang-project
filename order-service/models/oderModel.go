package models

import (
	// "time"

	// "go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// type OrderItem struct {
// 	ProductID		primitive.ObjectID		`bson:"product_id" validate:"required"`
// 	Name 			string					`bson:"name" validate:"required"`
// 	Quantity		int						`bson:"quantity" validate:"required"`
// 	Price 			float64					`bson:"price" validate:"required"`
// }

// type Order struct {
// 	ID 				primitive.ObjectID		`bson:"_id" json:"id"`
// 	UserID			primitive.ObjectID		`bson:"user_id" validate:"required"`
// 	Items			[]OrderItem				`bson:"items" validate:"required,dive"`
// 	TotalPrice 		float64					`bson:"total_price" validate:"required"`
// 	Status			string					`bson:"status" validate:"required"`
// 	Source			string					`bson:"source" validate:"required"`
// 	PaymentMethod	string					`bson:"payment_method" validate:"required"`
// 	PaymentStatus	string					`bson:"payment_status" validate:"required"`
// 	ShippingAddress	string					`bson:"shipping_address" validate:"required"`
// 	ShippingStatus	string					`bson:"shipping_status"`
// 	Created_at		time.Time				`bson:"created_at"`
// 	Updated_at		time.Time				`bson:"updated_at"`
// }

type Order struct {
	gorm.Model
	OrderID         string         `gorm:"uniqueIndex;not null"`
	UserID          string         `gorm:"not null"`
	Items           datatypes.JSON `gorm:"type:jsonb;not null"`
	Status          string         `gorm:"not null;default:'pending'"`
	Source          string         `gorm:"not null;default:'web'"`
	TotalPrice      float64        `gorm:"not null"`
	TotalCost       float64        `gorm:"not null;default:0"`
	TotalRevenue    float64        `gorm:"not null;default:0"`
	PaymentMethod   string         `gorm:"not null;default:'cod'"`
	PaymentStatus   string         `gorm:"not null;default:'unpaid'"`
	PaymentIntentID *string        `gorm:"column:payment_intent_id" json:"payment_intent_id,omitempty"`
	ShippingStatus  string         `gorm:"not null;default:'pending'"`
	ShippingAddress string         `gorm:"not null"`
	ShippingInfo    datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	// VendorID           *string        `gorm:"column:vendor_id" json:"vendor_id,omitempty"`
	PlatformFee        float64    `gorm:"not null;default:0"`
	VendorAmount       float64    `gorm:"not null;default:0"`
	DeliveryDate       *time.Time `json:"delivery_date"`
	PaymentReleaseDate *time.Time `json:"payment_release_date"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	VariantID string  `json:"variant_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	CostPrice float64 `json:"cost_price"`
	Price     float64 `json:"price"`
	VendorID  string  `json:"vendor_id"`
}

type TopProduct struct {
	ProductID     string  `json:"product_id"`
	Name          string  `json:"name"`
	TotalQuantity int64   `json:"total_quantity"`
	TotalSales    float64 `json:"total_sales"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalOrders   int64   `json:"total_orders"`
}

type MonthRevenue struct {
	Year    int     `json:"year"`
	Month   int     `json:"month"`
	Revenue float64 `json:"revenue"`
}

type TopCustomer struct {
	UserID        string    `json:"user_id"`
	TotalOrders   int64     `json:"total_orders"`
	TotalSpent    float64   `json:"total_spent"`
	TotalRevenue  float64   `json:"total_revenue"`
	LastOrderDate time.Time `json:"last_order_date"`
}

type SlowMovingProduct struct {
	ProductID         string     `json:"product_id"`
	Name              string     `json:"name"`
	LastSoldDate      *time.Time `json:"last_sold_date"`
	DaysSinceLastSale int        `json:"days_since_last_sale"`
	TotalStock        int        `json:"total_stock"`
}
