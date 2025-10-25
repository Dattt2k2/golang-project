package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type RatingConsumer struct {
	reader  *kafka.Reader
	handler RatingMessageHandler
}

type RatingMessageHandler interface {
	HandleRatingUpdate(ctx context.Context, message RatingUpdateMessage) error
}

type RatingUpdateMessage struct {
	ProductID          string   `json:"product_id"`
	NewReviewsCount    int      `json:"new_reviews_count"`
	NewReviewsSum      float64  `json:"new_reviews_sum"`
	ProcessedReviewIDs []string `json:"processed_review_ids"`
	Timestamp          string   `json:"timestamp"`
}

func NewRatingConsumer(brokers []string, topic, groupID string, handler RatingMessageHandler) *RatingConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	return &RatingConsumer{
		reader:  reader,
		handler: handler,
	}
}

func (c *RatingConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka consumer for rating updates...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka rating consumer...")
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading rating message: %v", err)
				continue
			}

			var ratingUpdate RatingUpdateMessage
			if err := json.Unmarshal(msg.Value, &ratingUpdate); err != nil {
				log.Printf("Error unmarshaling rating message: %v", err)
				continue
			}

			log.Printf("Received rating update for product %s: %d new reviews, sum: %.2f",
				ratingUpdate.ProductID, ratingUpdate.NewReviewsCount, ratingUpdate.NewReviewsSum)

			if err := c.handler.HandleRatingUpdate(ctx, ratingUpdate); err != nil {
				log.Printf("Error handling rating update: %v", err)
				continue
			}

			log.Printf("Successfully processed rating update for product %s", ratingUpdate.ProductID)
		}
	}
}

func (c *RatingConsumer) Close() error {
	return c.reader.Close()
}
