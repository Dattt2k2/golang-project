// package models

// import (
// 	"context"
// 	"time"

// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// // type Product struct{
// // 	ID            	primitive.ObjectID 		`bson:"_id,omitempty" json:"id,omitempty"`
// // 	Name			*string					`json:"name" validate:"required,min=2,max=100"`
// // 	Image_id		primitive.ObjectID		`bson:"image_id" json:"image_id"`
// // 	Description		*string					`json:"description" validate:"required,min=2,max=100"`
// // 	Quantity		*int					`json:"quantity" validate:"required,min=1"`
// // 	Price			float64					`json:"price" validate:"required"`
// // 	Created_at		time.Time				`json:"created_at"`
// // 	Updated_at		time.Time				`json:"updated_at"`
// // 	UserID			primitive.ObjectID		`json:"user_id"`
// // 	// ImageBase64		string					`json:"image,omitempty"`

// // }

// // Product - Database model (không có validation tags)
// type Product struct {
// 	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
// 	Name        string             `json:"name" bson:"name"`
// 	ImagePath   string             `json:"image_path" bson:"image_path"`
// 	Category    string             `json:"category" bson:"category"`
// 	Description string             `json:"description" bson:"description"`
// 	Quantity    int                `json:"quantity" bson:"quantity"`
// 	Price       float64            `json:"price" bson:"price"`
// 	SoldCount   int                `json:"sold_count" bson:"sold_count"`
// 	Created_at  time.Time          `json:"created_at" bson:"created_at"`
// 	Updated_at  time.Time          `json:"updated_at" bson:"updated_at"`
// 	UserID      string             `json:"user_id" bson:"user_id"`
// }

// // CreateProductRequest - Request struct cho tạo product mới
// type CreateProductRequest struct {
// 	Name        string  `json:"name" binding:"required,min=2,max=100"`
// 	ImagePath   string  `json:"image_path,omitempty"` // Optional - có thể empty hoặc có URL từ presigned upload
// 	Category    string  `json:"category" binding:"required"`
// 	Description string  `json:"description" binding:"required,min=2,max=500"`
// 	Quantity    int     `json:"quantity" binding:"required,min=1"`
// 	Price       float64 `json:"price" binding:"required,gt=0"`
// }

// // CreateProductWithImageRequest - Request struct khi upload ảnh cùng lúc
// type CreateProductWithImageRequest struct {
// 	Name        string  `json:"name" binding:"required,min=2,max=100"`
// 	Category    string  `json:"category" binding:"required"`
// 	Description string  `json:"description" binding:"required,min=2,max=500"`
// 	Quantity    int     `json:"quantity" binding:"required,min=1"`
// 	Price       float64 `json:"price" binding:"required,gt=0"`
// 	// Image sẽ được handle qua multipart form file
// }

// // UpdateProductRequest - Request struct cho update product
// type UpdateProductRequest struct {
// 	Name        *string  `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
// 	ImagePath   *string  `json:"image_path,omitempty"` // Optional update
// 	Category    *string  `json:"category,omitempty"`
// 	Description *string  `json:"description,omitempty" binding:"omitempty,min=2,max=500"`
// 	Quantity    *int     `json:"quantity,omitempty" binding:"omitempty,min=1"`
// 	Price       *float64 `json:"price,omitempty" binding:"omitempty,gt=0"`
// }

// // ProductResponse - Response struct cho API
// type ProductResponse struct {
// 	ID          string    `json:"id"`
// 	Name        string    `json:"name"`
// 	ImagePath   string    `json:"image_path"`
// 	Category    string    `json:"category"`
// 	Description string    `json:"description"`
// 	Quantity    int       `json:"quantity"`
// 	Price       float64   `json:"price"`
// 	SoldCount   int       `json:"sold_count"`
// 	Created_at  time.Time `json:"created_at"`
// 	Updated_at  time.Time `json:"updated_at"`
// }

// type StockUpdateItem struct {
// 	ProductID string
// 	Quantity  int
// }

// type ProductStockUpdater interface {
// 	UpdateProductStock(ctx context.Context, id primitive.ObjectID, quantity int) error
// 	IncrementSoldCount(ctx context.Context, productID string, quantity int) error
// 	DecrementSoldCount(ctx context.Context, productID string, quantity int) error
// }

// // PresignedUploadRequest - Request để lấy presigned URL
// type PresignedUploadRequest struct {
// 	Filename    string `json:"filename" binding:"required"`
// 	ContentType string `json:"content_type" binding:"required"`
// }

// // PresignedUploadResponse - Response chứa presigned URL
// type PresignedUploadResponse struct {
// 	PresignedURL string `json:"presigned_url"` // URL để upload
// 	PublicURL    string `json:"public_url"`    // URL sau khi upload thành công
// 	ExpiresIn    int    `json:"expires_in"`    // Thời gian hết hạn (seconds)
// }


package models

import (
    "context"
    "time"
)

// Product - Database model for DynamoDB
type Product struct {
    ID          string    `json:"id" dynamodbav:"id"`                         // Thay đổi từ ObjectID sang string
    Name        string    `json:"name" dynamodbav:"name"`
    ImagePath   string    `json:"image_path" dynamodbav:"image_path"`
    Category    string    `json:"category" dynamodbav:"category"`
    Description string    `json:"description" dynamodbav:"description"`
    Quantity    int       `json:"quantity" dynamodbav:"quantity"`
    Price       float64   `json:"price" dynamodbav:"price"`
    SoldCount   int       `json:"sold_count" dynamodbav:"sold_count"`
    Created_at  time.Time `json:"created_at" dynamodbav:"created_at"`
    Updated_at  time.Time `json:"updated_at" dynamodbav:"updated_at"`
    UserID      string    `json:"user_id" dynamodbav:"user_id"`
}

// CreateProductRequest - Request struct cho tạo product mới
type CreateProductRequest struct {
    VendorID   string  `json:"vendor_id" binding:"required"` // Thêm trường vendor_id
    Name        string  `json:"name" binding:"required,min=2,max=100"`
    ImagePath   string  `json:"image_path,omitempty"` // Optional - có thể empty hoặc có URL từ presigned upload
    Category    string  `json:"category" binding:"required"`
    Description string  `json:"description" binding:"required,min=2,max=500"`
    Quantity    int     `json:"quantity" binding:"required,min=1"`
    Price       float64 `json:"price" binding:"required,gt=0"`
}

// CreateProductWithImageRequest - Request struct khi upload ảnh cùng lúc
type CreateProductWithImageRequest struct {
    VendorID    string  `json:"vendor_id" binding:"required"` // Thêm trường vendor_id
    Name        string  `json:"name" binding:"required,min=2,max=100"`
    Category    string  `json:"category" binding:"required"`
    Description string  `json:"description" binding:"required,min=2,max=500"`
    Quantity    int     `json:"quantity" binding:"required,min=1"`
    Price       float64 `json:"price" binding:"required,gt=0"`
    // Image sẽ được handle qua multipart form file
}

// UpdateProductRequest - Request struct cho update product
type UpdateProductRequest struct {
    Name        *string  `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
    ImagePath   *string  `json:"image_path,omitempty"` // Optional update
    Category    *string  `json:"category,omitempty"`
    Description *string  `json:"description,omitempty" binding:"omitempty,min=2,max=500"`
    Quantity    *int     `json:"quantity,omitempty" binding:"omitempty,min=1"`
    Price       *float64 `json:"price,omitempty" binding:"omitempty,gt=0"`
}

// ProductResponse - Response struct cho API
type ProductResponse struct {
    ID          string    `json:"id"`
    VendorID    string    `json:"vendor_id"` 
    Name        string    `json:"name"`
    ImagePath   string    `json:"image_path"`
    Category    string    `json:"category"`
    Description string    `json:"description"`
    Quantity    int       `json:"quantity"`
    Price       float64   `json:"price"`
    SoldCount   int       `json:"sold_count"`
    Created_at  time.Time `json:"created_at"`
    Updated_at  time.Time `json:"updated_at"`
}

type StockUpdateItem struct {
    ProductID string  // Thay đổi từ primitive.ObjectID sang string
    Quantity  int
}

type ProductStockUpdater interface {
    UpdateProductStock(ctx context.Context, id string, quantity int) error        // Thay đổi parameter type
    IncrementSoldCount(ctx context.Context, productID string, quantity int) error
    DecrementSoldCount(ctx context.Context, productID string, quantity int) error
}

// PresignedUploadRequest - Request để lấy presigned URL
type PresignedUploadRequest struct {
    Filename    string `json:"filename" binding:"required"`
    ContentType string `json:"content_type" binding:"required"`
}

// PresignedUploadResponse - Response chứa presigned URL
type PresignedUploadResponse struct {
    PresignedURL string `json:"presigned_url"` // URL để upload
    PublicURL    string `json:"public_url"`    // URL sau khi upload thành công
    ExpiresIn    int    `json:"expires_in"`    // Thời gian hết hạn (seconds)
}