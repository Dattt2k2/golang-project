package cron

import (
	"context"
	"fmt"
	"log"
	"time"

	"review-service/internal/kafka"
	"review-service/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type PendingReviewAggregator struct {
	dynamoDB           *dynamodb.Client
	kafkaProducer      *kafka.Producer
	reviewPendingTable string
}

func NewPendingReviewAggregator(db *dynamodb.Client, producer *kafka.Producer, reviewPendingTable string) *PendingReviewAggregator {
	return &PendingReviewAggregator{
		dynamoDB:           db,
		kafkaProducer:      producer,
		reviewPendingTable: reviewPendingTable,
	}
}

func (pra *PendingReviewAggregator) Run(ctx context.Context) error {
	log.Println("Starting pending review aggregation job...")
	startTime := time.Now()

	// 1. Lấy tất cả reviews từ bảng review_pending
	pendingReviews, err := pra.getAllPendingReviews(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending reviews: %w", err)
	}

	if len(pendingReviews) == 0 {
		log.Println("No pending reviews to process")
		return nil
	}

	log.Printf("Found %d pending reviews", len(pendingReviews))

	// 2. Group theo ProductID và tính tổng
	aggregates := pra.aggregateByProduct(pendingReviews)
	log.Printf("Aggregated into %d products", len(aggregates))

	// 3. Gửi từng aggregate lên Kafka và xóa reviews đã xử lý
	successCount := 0
	failCount := 0
	deletedCount := 0

	for _, aggregate := range aggregates {
		message := models.RatingUpdateMessage{
			ProductID:          aggregate.ProductID,
			NewReviewsCount:    aggregate.TotalReviews,
			NewReviewsSum:      aggregate.SumRating,
			ProcessedReviewIDs: aggregate.ReviewIDs,
			Timestamp:          time.Now().Format(time.RFC3339),
		}

		// Gửi message lên Kafka
		err := pra.kafkaProducer.PublishRatingUpdate(ctx, message)
		if err != nil {
			log.Printf("Failed to publish update for product %s: %v", aggregate.ProductID, err)
			failCount++
			continue
		}

		// 4. XÓA các reviews đã xử lý khỏi review_pending
		deleted, err := pra.deleteProcessedReviews(ctx, aggregate.ReviewIDs)
		if err != nil {
			log.Printf("Failed to delete reviews for product %s: %v", aggregate.ProductID, err)
			continue
		}

		deletedCount += deleted
		successCount++
		log.Printf("Processed product %s: sent to Kafka and deleted %d reviews", aggregate.ProductID, deleted)
	}

	duration := time.Since(startTime)
	log.Printf("Pending review aggregation completed in %v. Success: %d, Failed: %d, Deleted: %d reviews",
		duration, successCount, failCount, deletedCount)

	return nil
}

func (pra *PendingReviewAggregator) getAllPendingReviews(ctx context.Context) ([]models.PendingReview, error) {
	// Scan toàn bộ bảng review_pending
	input := &dynamodb.ScanInput{
		TableName: aws.String(pra.reviewPendingTable),
	}

	result, err := pra.dynamoDB.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan review_pending table: %w", err)
	}

	var reviews []models.PendingReview
	err = attributevalue.UnmarshalListOfMaps(result.Items, &reviews)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal reviews: %w", err)
	}

	return reviews, nil
}

func (pra *PendingReviewAggregator) aggregateByProduct(reviews []models.PendingReview) []models.PendingReviewAggregate {
	aggregateMap := make(map[string]*models.PendingReviewAggregate)

	for _, review := range reviews {
		if agg, exists := aggregateMap[review.ProductID]; exists {
			agg.TotalReviews++
			agg.SumRating += review.Rating
			agg.ReviewIDs = append(agg.ReviewIDs, review.ReviewID)
		} else {
			aggregateMap[review.ProductID] = &models.PendingReviewAggregate{
				ProductID:    review.ProductID,
				TotalReviews: 1,
				SumRating:    review.Rating,
				ReviewIDs:    []string{review.ReviewID},
			}
		}
	}

	aggregates := make([]models.PendingReviewAggregate, 0, len(aggregateMap))
	for _, agg := range aggregateMap {
		aggregates = append(aggregates, *agg)
	}

	return aggregates
}

func (pra *PendingReviewAggregator) deleteProcessedReviews(ctx context.Context, reviewIDs []string) (int, error) {
	deletedCount := 0

	for _, reviewID := range reviewIDs {
		input := &dynamodb.DeleteItemInput{
			TableName: aws.String(pra.reviewPendingTable),
			Key: map[string]types.AttributeValue{
				"review_id": &types.AttributeValueMemberS{
					Value: reviewID,
				},
			},
		}

		_, err := pra.dynamoDB.DeleteItem(ctx, input)
		if err != nil {
			log.Printf("Failed to delete review %s: %v", reviewID, err)
			continue
		}
		deletedCount++
	}

	log.Printf("Successfully deleted %d/%d reviews from review_pending", deletedCount, len(reviewIDs))
	return deletedCount, nil
}
