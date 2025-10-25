package handler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"product-service/kafka"
)

type RatingUpdateHandler struct {
	dynamoDB     *dynamodb.DynamoDB
	productTable string
}

type Product struct {
	ProductID   string  `json:"product_id" dynamodbav:"product_id"`
	Rating      float64 `json:"rating" dynamodbav:"rating"`
	ReviewCount int     `json:"review_count" dynamodbav:"review_count"`
	// ... other fields
}

func NewRatingUpdateHandler(db *dynamodb.DynamoDB, productTable string) *RatingUpdateHandler {
	return &RatingUpdateHandler{
		dynamoDB:     db,
		productTable: productTable,
	}
}
func (h *RatingUpdateHandler) HandleRatingUpdate(ctx context.Context, msg kafka.RatingUpdateMessage) error {
	if msg.ProductID == "" {
		return fmt.Errorf("empty product_id in rating message")
	}
	if msg.NewReviewsCount <= 0 {
		log.Printf("Skipping rating update for product %s: new review count %d", msg.ProductID, msg.NewReviewsCount)
		return nil
	}
	if msg.NewReviewsSum < 0 {
		return fmt.Errorf("invalid new reviews sum: %.2f", msg.NewReviewsSum)
	}

	log.Printf("Received rating update: product=%s count=%d sum=%.2f",
		msg.ProductID, msg.NewReviewsCount, msg.NewReviewsSum)

	if err := h.updateProductRating(ctx, msg.ProductID, msg.NewReviewsSum, msg.NewReviewsCount, time.Now().Format(time.RFC3339)); err != nil {
		log.Printf("Error applying rating update for product %s: %v", msg.ProductID, err)
		return err
	}

	log.Printf("Successfully applied rating update for product %s", msg.ProductID)
	return nil
}

func (h *RatingUpdateHandler) updateProductRating(ctx context.Context, productID string, rating float64, reviewCount int, timestamp string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(h.productTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(productID),
			},
		},
		UpdateExpression: aws.String("SET rating = :rating, review_count = :count, updated_at = :timestamp"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":rating": {
				N: aws.String(fmt.Sprintf("%.2f", rating)),
			},
			":count": {
				N: aws.String(fmt.Sprintf("%d", reviewCount)),
			},
			":timestamp": {
				S: aws.String(timestamp),
			},
		},
	}

	_, err := h.dynamoDB.UpdateItemWithContext(ctx, input)
	return err
}
