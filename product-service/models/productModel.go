package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// type Product struct{
// 	ID            	primitive.ObjectID 		`bson:"_id,omitempty" json:"id,omitempty"`
// 	Name			*string					`json:"name" validate:"required,min=2,max=100"`
// 	Image_id		primitive.ObjectID		`bson:"image_id" json:"image_id"`
// 	Description		*string					`json:"description" validate:"required,min=2,max=100"`
// 	Quantity		*int					`json:"quantity" validate:"required,min=1"`
// 	Price			float64					`json:"price" validate:"required"`
// 	Created_at		time.Time				`json:"created_at"`
// 	Updated_at		time.Time				`json:"updated_at"`
// 	UserID			primitive.ObjectID		`json:"user_id"`
// 	// ImageBase64		string					`json:"image,omitempty"`

// }


type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        *string            `json:"name" validate:"required,min=2,max=100"`
	ImagePath   string             `json:"image_path"` 
	Category	*string             `json:"category validate:"required"`
	Description *string            `json:"description" validate:"required,min=2,max=100"`
	Quantity    *int               `json:"quantity" validate:"required,min=1"`
	Price       float64            `json:"price" validate:"required"`
	Created_at  time.Time          `json:"created_at"`
	Updated_at  time.Time          `json:"updated_at"`
	UserID      primitive.ObjectID `json:"user_id"`
}