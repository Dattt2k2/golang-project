package models

import (
	"context"
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
	SoldCount   *int               `json:"sold_count" validate:"required,default=0"`
	Created_at  time.Time          `json:"created_at"`
	Updated_at  time.Time          `json:"updated_at"`
	UserID      primitive.ObjectID `json:"user_id"`
}


type StockUpdateItem struct {
    ProductID string
    Quantity  int
}

type ProductStockUpdater interface {
	UpdateProductStock(ctx context.Context, id primitive.ObjectID, quantity int) error
    IncrementSoldCount(ctx context.Context, productID string, quantity int) error
    DecrementSoldCount(ctx context.Context, productID string, quantity int) error
}