package models

type PendingReview struct {
	ReviewID  string  `json:"review_id" dynamodbav:"review_id"`
	ProductID string  `json:"product_id" dynamodbav:"product_id"`
	Rating    float64 `json:"rating" dynamodbav:"rating"`
	CreatedAt string  `json:"created_at" dynamodbav:"created_at"`
}

type PendingReviewAggregate struct {
	ProductID    string   `json:"product_id"`
	TotalReviews int      `json:"total_reviews"`
	SumRating    float64  `json:"sum_rating"`
	ReviewIDs    []string `json:"review_ids"` // Để xóa sau khi xử lý
}

type RatingUpdateMessage struct {
	ProductID          string   `json:"product_id"`
	NewReviewsCount    int      `json:"new_reviews_count"`
	NewReviewsSum      float64  `json:"new_reviews_sum"`
	ProcessedReviewIDs []string `json:"processed_review_ids"`
	Timestamp          string   `json:"timestamp"`
}
