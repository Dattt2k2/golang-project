package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartItem struct{
	ProductID			primitive.ObjectID		`bson:"product_id" validate:"required"`
	Quantity			int						`bson:"quantity" validate:"required,min=1"`
	Price       		float64            		`json:"price" validate:"required"`
	Name           		string					`json:"name" validate:"required"`
	ImageUrl			string					`json:"image_url" validate:"required"`
	Description			string					`json:"description" validate"required`
}

type Cart struct{
	ID					primitive.ObjectID		`bson:"_id, omitempty"`
	userId				primitive.ObjectID		`json:"user_id"`
	Items				[]CartItem				`json:"items"`
	Created_at			time.Time				`json:"created_at"`
	Updated_at			time.Time				`json:"updated_at"`
}