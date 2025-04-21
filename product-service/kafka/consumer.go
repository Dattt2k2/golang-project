package kafka

import (
	"context"
	"encoding/json"
	"log"

	controller "github.com/Dattt2k2/golang-project/product-service/controller"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	OrderSuccessTopic = "order_success"
	OrderReturnedTopic = "order_returned"
)

type OrderSuccessEvent struct {
	OrderID    string `json:"order_id"`
	UserID    string `json:"user_id"`
	Items      []OrderItemInfo `json:"items"`
	TotalPrice float64 `json:"total_price"`
}

type OrderItemInfo struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Price     float64 `json:"price"`
}

func ConsumeOrderSuccess(brokers []string, productCtrl controller.ProductController){
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   OrderSuccessTopic,
		GroupID: "product-service",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	go func() {
		for {
			message, err := reader.ReadMessage(context.Background())
			if err != nil{
				log.Printf("Error reading message: %v", err)
				continue
			}

			var event OrderSuccessEvent
			if err := json.Unmarshal(message.Value, &event); err != nil{
				log.Printf("Error unmarshalling message: %v", err)
				continue
			}

			userID, err := primitive.ObjectIDFromHex(event.UserID)
			if err != nil{
				log.Printf("Error converting user ID: %v", err)
				continue 
			}

			// Convert OrderItemInfo to controller.StockUpdateItem
			stockItems := make([]controller.StockUpdateItem, len(event.Items))
			for i, item := range event.Items {
				stockItems[i] = controller.StockUpdateItem{
					ProductID: item.ProductID,
					Quantity:  item.Quantity,
				}
			}
			
			if err := productCtrl.UpdateProductStock(context.Background(), stockItems, false); err != nil{
				log.Printf("Error updating product stock: %v", err)
				continue 
			}

			log.Printf("Product stock updated successfully for user: %v", userID)
		}
	}()

	log.Printf("Kafka consumer started for topic: %s", OrderSuccessTopic)
}


type OrderReturnedEvent struct {
	OrderID    string `json:"order_id"`
	UserID    string `json:"user_id"`
	Items      []OrderItemInfo `json:"items"`
	TotalPrice float64 `json:"total_price"`
}

func ConsumerOrderReturned(brokers []string, productCtrl controller.ProductController) {
	reader := kafka.NewReader (kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   OrderReturnedTopic,
		GroupID: "product-service",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB	
	})

	go func() {
		for {
			message, err := reader.ReadMessage(context.Background())
			if err != nil{
				log.Printf("Error reading message: %v", err)
				continue
			}
			var event OrderReturnedEvent
			if err := json.Unmarshal(message.Value, &event); err != nil{
				log.Printf("Error unmarshalling message: %v", err)
				continue
			}

			userID, err := primitive.ObjectIDFromHex(event.UserID)
			if err != nil{
				log.Printf("Error converting user ID: %v", err)
				continue 
			}

			// Convert OrderItemInfo to controller.StockUpdateItem
			stockItems := make([]controller.StockUpdateItem, len(event.Items))
			for i, item := range event.Items {
				stockItems[i] = controller.StockUpdateItem{
					ProductID: item.ProductID,
					Quantity:  item.Quantity,
				}
			}
			
			if err := productCtrl.UpdateProductStock(context.Background(), stockItems, true); err != nil{
				log.Printf("Error updating product stock: %v", err)
				continue 
			}

			log.Printf("Product stock updated successfully for user: %v", userID)
		}
		
	}()
	log.Printf("Kafka consumer started for topic: %s", OrderReturnedTopic)
}