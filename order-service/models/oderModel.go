package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	ProductID		primitive.ObjectID		`bson:"product_id" validate:"required"`
	Name 			string					`bson:"name" validate:"required"`
	Quantity		int						`bson:"quantity" validate:"required"`
	Price 			float64					`bson:"price" validate:"required"`
}

type Order struct {
	ID 				primitive.ObjectID		`bson:"_id" json:"id"`
	UserID			primitive.ObjectID		`bson:"user_id" validate:"required"`
	Items			[]OrderItem				`bson:"items" validate:"required,dive"`
	TotalPrice 		float64					`bson:"total_price" validate:"required"`
	Status			string					`bson:"status" validate:"required"`
	Source			string					`bson:"source" validate:"required"`
	PaymentMethod	string					`bson:"payment_method" validate:"required"`
	PaymentStatus	string					`bson:"payment_status" validate:"required"`
	ShippingAddress	string					`bson:"shipping_address" validate:"required"`
	Created_at		time.Time				`bson:"created_at"`
	Updated_at		time.Time				`bson:"updated_at"`
}

