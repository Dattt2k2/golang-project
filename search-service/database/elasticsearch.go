package database

import (
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"search-service/log"
)

var ES *elasticsearch.Client


func InitElasticsearch() {

	elasticsearchURL := os.Getenv("ELASTICSEARCH_URL")

	cfg := elasticsearch.Config{
		Addresses: []string{elasticsearchURL},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		logger.Err("Error creating Elasticsearch client", err)
	}

	ES = client
}