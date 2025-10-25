package models

import "time"

type Review struct {
	ID        string    `json:"id" dynamodbav:"id"`
	ProductID string    `json:"product_id" dynamodbav:"product_id"`
	UserID    string    `json:"user_id" dynamodbav:"user_id"`
	Rating    int       `json:"rating" binding:"required,min=1,max=5" dynamodbav:"rating"`
	Title     string    `json:"title,omitempty" dynamodbav:"title"`
	Body      string    `json:"body_review,omitempty" dynamodbav:"body"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
}

type SumReviewPending struct {
	ProductID   string    `json:"product_id" dynamodbav:"product_id"`
	ReviewID    string    `json:"review_id", dynamodbav:"review_id"`
	LastTimeSum time.Time `json:"last_time_sum" dynamodbav:"last_time_sum"`
	Rating      int       `json:"rating" dynamodbav:"rating"`
}
