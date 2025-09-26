// package models

// import (
// 	"time"

// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// type CartItem struct{
// 	ProductID			primitive.ObjectID		`bson:"product_id" validate:"required"`
// 	Quantity			int						`bson:"quantity" validate:"required,min=1"`
// 	Price       		float64            		`json:"price" validate:"required"`
// 	Name           		string					`json:"name" validate:"required"`
// 	ImageUrl			string					`json:"image_url" validate:"required"`
// 	Description			string					`json:"description" validate"required`
// }

// type Cart struct{
// 	ID					primitive.ObjectID		`bson:"_id, omitempty"`
// 	userId				primitive.ObjectID		`json:"user_id"`
// 	Items				[]CartItem				`json:"items"`
// 	Created_at			time.Time				`json:"created_at"`
// 	Updated_at			time.Time				`json:"updated_at"`
// }

package models

import (
    "time"
)

type CartItem struct {
    VendorID    string  `json:"vendor_id" dynamodbav:"vendor_id" validate:"required"`
    ProductID   string  `json:"product_id" dynamodbav:"product_id" validate:"required"`
    Quantity    int     `json:"quantity" dynamodbav:"quantity" validate:"required,min=1"`
    Price       float64 `json:"price" dynamodbav:"price" validate:"required"`
    Name        string  `json:"name" dynamodbav:"name" validate:"required"`
    ImageUrl    string  `json:"image_url" dynamodbav:"image_url" validate:"required"`
    Description string  `json:"description" dynamodbav:"description" validate:"required"`
}

type Cart struct {
    ID         string     `json:"id" dynamodbav:"id"`
    UserID     string     `json:"user_id" dynamodbav:"user_id"`
    Items      []CartItem `json:"items" dynamodbav:"items"`
    Created_at time.Time  `json:"created_at" dynamodbav:"created_at"`
    Updated_at time.Time  `json:"updated_at" dynamodbav:"updated_at"`
}