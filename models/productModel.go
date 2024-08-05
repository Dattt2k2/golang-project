package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct{
	ID             primitive.ObjectID 		`bson:"_id,omitempty" json:"id,omitempty"`
	Name			string					`json:"name"`
	Image			string					`json:"image"`
	Description		string					`json:"description"`
	Price			float64					`json:"price"`
	Created_at		time.Time				`json:"created_at"`
	Updated_at		time.Time				`json:"updated_at"`
}