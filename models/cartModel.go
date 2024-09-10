package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartItem struct{
	ProductID			primitive.ObjectID		`bson:"_id" validate:"requried"`
	Quantity			*int					`bson:"quantity" validate:"required,min=1"`
}

type Cart struct{
	ID					primitive.ObjectID		`bson:"_id, omitempty"`
	userId				primitive.ObjectID		`json:"user_id"`
	Items				[]CartItem				`json:"items"`
	Created_at			time.Time				`json:"created_at"`
	Updated_at			time.Time				`json:"updated_at"`
}