package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Dattt2k2/golang-project/product-service/models"
	"github.com/Dattt2k2/golang-project/product-service/log"
	"github.com/segmentio/kafka-go"
) 


const (
	ProductEventTopic = "product-events" 
)

var (
	productEventWriter *kafka.Writer
)


type ProductEvent struct {
	Type string `json:"type"` 
	Product *models.Product `json:"product"`
	ID string `json:"id"`
}

func InitProductEventProducer(brokers []string) {
	productEventWriter = &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    ProductEventTopic,
		Balancer: &kafka.LeastBytes{},
	}
	logger.Logger.Infof("Product event producer initialized with brokers: %v", brokers)
}

func ProduceProductEvent(ctx context.Context, eventType string, product *models.Product, id string) error  {
	if productEventWriter == nil {
		logger.Err("Product event writer is not initialized", nil)
		return fmt.Errorf("product event writer is not initialized")
	}

	event := ProductEvent {
		Type : eventType,
		Product : product,
		ID : id,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		logger.Err("Failed to marshal product event", err)
		return err 
	}

	message := kafka.Message {
		Key : []byte(id),
		Value : payload,
	}

	if err := productEventWriter.WriteMessages(ctx, message); err != nil {
		logger.Err("Failed to write product event message", err)
		return err 
	}

	logger.Logger.Infof("Product event produced: %s", eventType)
	return nil 
}


func CloseProductEventProducer() {
	if productEventWriter != nil {
		logger.Info("Closing product event writer")
		productEventWriter.Close()
	}
}



