package models

import "time"

type ProductVariant struct {
	ID        string    `json:"id"`
	Size      string    `json:"size"`
	Color     string    `json:"color"`
	Material  string    `json:"material"`
	CostPrice float64   `json:"cost_price"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

type Product struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	ImagePath   []string         `json:"image_path"`
	Category    string           `json:"category"`
	Description string           `json:"description"`
	Variants    []ProductVariant `json:"variants"`
	SoldCount   int              `json:"sold_count"`
	Created_at  time.Time        `json:"created_at"`
	Updated_at  time.Time        `json:"updated_at"`
	UserID      string           `json:"user_id"`
	Status      string           `json:"status"`
	Rating      float64          `json:"rating"`
	RatingCount int              `json:"rating_count"`
}

type PresignedUploadRequest struct {
	Filename    string `json:"fileName"`
	ContentType string `json:"fileType,omitempty"`
}

// PresignedUploadResponse - Response chứa presigned URL
type PresignedUploadResponse struct {
	PresignedURL string `json:"presigned_url"`
	S3Key        string `json:"s3_key"`
	Filename     string `json:"filename"`
	ExpiresAt    int64  `json:"expires_at"`
	ExpiresIn    int    `json:"expires_in"`
}

type AdvancedSearchResponse struct {
	Data      []Product              `json:"data"`
	Total     int                    `json:"total"`
	HavePrev  bool                   `json:"has_prev"`
	HaveNext  bool                   `json:"has_next"`
	Filters   map[string]interface{} `json:"filters"`
	From      int                    `json:"from"`
	Limit     int                    `json:"limit"`
	Page      int                    `json:"page"`
	Query     string                 `json:"query"`
	SortBy    string                 `json:"sortBy"`
	SortOrder string                 `json:"sortOrder"`
}
