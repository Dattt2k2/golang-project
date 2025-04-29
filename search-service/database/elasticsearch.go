package database

import (
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

var ES *elasticsearch.Client


func InitElasticsearch() {

	elasticsearchURL := os.Getenv("ELASTICSEARCH_URL")

	cfg := elasticsearch.Config{
		Addresses: []string{elasticsearchURL},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	ES = client
}