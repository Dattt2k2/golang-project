package events

import (
	"encoding/json"
	"log"
)

// EventPublisher publishes domain events to an external system (kafka, rabbitmq, etc.).
// This is a simple logging publisher used as a stub; swap with a real implementation later.
type EventPublisher interface {
	Publish(topic string, payload interface{}) error
}

type LoggingPublisher struct{}

func NewLoggingPublisher() *LoggingPublisher { return &LoggingPublisher{} }

func (p *LoggingPublisher) Publish(topic string, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	log.Printf("[event] topic=%s payload=%s\n", topic, string(b))
	return nil
}
