package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"user-service/internal/models"
	"user-service/internal/repository"
	logger "user-service/log"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// StartUserCreatedConsumer starts a goroutine that listens for user.created events and persists them.
func StartUserCreatedConsumer(brokers []string, topic string, repo repository.UserRepository) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "user-service-consumer",
	})

	go func() {
		defer r.Close()
		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				log.Printf("user consumer read error: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
			var payload map[string]interface{}
			if err := json.Unmarshal(m.Value, &payload); err != nil {
				log.Printf("failed to unmarshal user.created: %v", err)
				continue
			}

			// Build user model (simple mapping, adjust as needed)
			var uid uuid.UUID
			if v, ok := payload["id"].(string); ok {
				if parsed, perr := uuid.Parse(v); perr == nil {
					uid = parsed
				} else {
					logger.Error("invalid uuid in payload.id")
					return 
				}
			} else {
				return 
			}
			email, _ := payload["email"].(string)
			firstName, _ := payload["first_name"].(string)
			lastName, _ := payload["last_name"].(string)
			phone, _ := payload["phone"].(string)
			userType, _ := payload["user_type"].(string)

			u := &models.User{
				ID:        uid,
				Email:     &email,
				FirstName: &firstName,
				LastName:  &lastName,
				Phone:     &phone,
				UserType:  &userType,
			}

			// Idempotency: check by email
			// repository can implement FindByEmail or use existing methods; we'll try Save and ignore unique constraint errors.
			if err := repo.SaveUser(u); err != nil {
				log.Printf("failed to save user from event: %v", err)
				continue
			}
			log.Printf("user created from event: %s", email)
		}
	}()
}
