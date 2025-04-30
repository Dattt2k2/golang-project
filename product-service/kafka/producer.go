package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Dattt2k2/golang-project/product-service/models"
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
	log.Printf("Product event producer initialized")
}

func ProduceProductEvent(ctx context.Context, eventType string, product *models.Product, id string) error  {
	if productEventWriter == nil {
		log.Printf("Product event writer is not initialized")
		return fmt.Errorf("product event writer is not initialized")
	}

	event := ProductEvent {
		Type : eventType,
		Product : product,
		ID : id,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal product event: %v", err)
		return err 
	}

	message := kafka.Message {
		Key : []byte(id),
		Value : payload,
	}

	if err := productEventWriter.WriteMessages(ctx, message); err != nil {
		log.Printf("Failed to write product event message: %v", err)
		return err 
	}

	log.Printf("Product event produced: %v", event)
	return nil 
}


func CloseProductEventProducer() {
	if productEventWriter != nil {
		productEventWriter.Close()
	}
}



