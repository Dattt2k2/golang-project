package handler

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

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

func (h *RatingUpdateHandler) HandleRatingUpdate(ctx context.Context, message kafka.RatingUpdateMessage) error {
	// 1. Lấy product hiện tại
	product, err := h.getProduct(ctx, message.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	// 2. Tính rating mới
	// Công thức: (rating_hiện_tại * số_review_hiện_tại + tổng_rating_mới) / (số_review_hiện_tại + số_review_mới)
	currentTotalRating := product.Rating * float64(product.ReviewCount)
	newTotalRating := currentTotalRating + message.NewReviewsSum
	newReviewCount := product.ReviewCount + message.NewReviewsCount
	newAverageRating := newTotalRating / float64(newReviewCount)

	log.Printf("Product %s: Current rating %.2f (%d reviews) -> New rating %.2f (%d reviews)",
		message.ProductID, product.Rating, product.ReviewCount, newAverageRating, newReviewCount)

	// 3. Update product
	err = h.updateProductRating(ctx, message.ProductID, newAverageRating, newReviewCount, message.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to update product rating: %w", err)
	}

	log.Printf("Successfully updated rating for product %s", message.ProductID)
	return nil
}

func (h *RatingUpdateHandler) getProduct(ctx context.Context, productID string) (*Product, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(h.productTable),
		Key: map[string]*dynamodb.AttributeValue{
			"product_id": {
				S: aws.String(productID),
			},
		},
	}

	result, err := h.dynamoDB.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, fmt.Errorf("product not found: %s", productID)
	}

	var product Product
	err = dynamodbattribute.UnmarshalMap(result.Item, &product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (h *RatingUpdateHandler) updateProductRating(ctx context.Context, productID string, rating float64, reviewCount int, timestamp string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(h.productTable),
		Key: map[string]*dynamodb.AttributeValue{
			"product_id": {
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
