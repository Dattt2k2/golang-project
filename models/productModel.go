package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct{
	ID             primitive.ObjectID 		`bson:"_id,omitempty" json:"id,omitempty"`
	Name			*string					`json:"name" validate:"required,min=2,max=100"`
	Image			primitive.Binary		`json:"image"`
	Description		*string					`json:"description" validate:"required,min=2,max=100"`
	Price			float64					`json:"price" validate:"required"`
	Created_at		time.Time				`json:"created_at"`
	Updated_at		time.Time				`json:"updated_at"`
	UserID			primitive.ObjectID		`json:"user_id"`
}