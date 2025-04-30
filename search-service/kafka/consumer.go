package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Dattt2k2/golang-project/search-service/models"
	"github.com/Dattt2k2/golang-project/search-service/repository"
	"github.com/segmentio/kafka-go"
)


type ProductEvent struct {
	Type string `json:"type"`
	Product *models.Product `json:"product"`
	ID string `json:"id"`
}

func InitProductEventConsumer(repo repository.SearchRepository,brokers []string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "product-events",
		GroupID: "search-service",
	})

	defer r.Close()

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue 
		}
		var event ProductEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue 
		}

		switch event.Type {
		case "created", "updated":
			if event.Product != nil {
				if repo.IndexProduct(event.Product) != nil {
					log.Printf("Error indexing product: %v", err)

				}
			}
		case "deleted":
			if event.ID != "" {
				if err := repo.DeleteProduct(event.ID); err != nil {
					log.Printf("Error deleting product: %v", err)
				}
			}
		}
	}
}