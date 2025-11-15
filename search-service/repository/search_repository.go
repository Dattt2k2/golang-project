package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"search-service/database"
	"search-service/models"
)

type SearchRepository interface {
	BasicSearch(query string) ([]models.Product, error ) 
	AdvancedSearch(query string, filters map[string]interface{}) ([]models.Product, error)
	IndexProduct(product *models.Product) error  
	DeleteProduct(id string) error 
}


type searchRepository struct {}

func NewSearchRepository() SearchRepository {
	return &searchRepository{}
}


func (r *searchRepository) BasicSearch(query string) ([]models.Product, error) {
	var buf bytes.Buffer
	esQuery := map[string]interface{}{
		"query":map[string]interface{}{
			"multi_match":map[string]interface{}{
				"query": query,
				"fields": []string{"name", "category"},
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(esQuery); err != nil {
		return nil, err 
	}
	res, err := database.ES.Search(
		database.ES.Search.WithContext(context.Background()),
		database.ES.Search.WithIndex(os.Getenv("ELASTICSEARCH_INDEX")),
		database.ES.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err 
	}

	defer res.Body.Close()
	
	var rResult struct {
		Hits struct {
			Hits []struct {
				Source models.Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&rResult); err != nil {
		return nil, err 
	}

	products := make([]models.Product, 0)
	for _, hit := range rResult.Hits.Hits {
		products = append(products, hit.Source)
	}
	return products, nil 
}

func (r *searchRepository) AdvancedSearch(query string, filters map[string]interface{}) ([]models.Product, error) {
	var buf bytes.Buffer

	boolQuery := map[string]interface{}{
		"must": []interface{}{
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  query,
					"fields": []string{"name", "description", "category"},
				},
			},
		},
		"filter": []interface{}{},
	}

	if price, ok := filters["price"]; ok {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"range": map[string]interface{}{
				"price": price,
			},
		})
	}

	if category, ok := filters["category"]; ok {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"category": category,
			},
		})
	}

	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQuery,
		},
	}

	if err := json.NewEncoder(&buf).Encode(esQuery); err != nil {
		return nil, err 
	}

	res, err := database.ES.Search(
		database.ES.Search.WithContext(context.Background()),
		database.ES.Search.WithIndex(os.Getenv("ELASTICSEARCH_INDEX")),
		database.ES.Search.WithBody(&buf),
	)

	if err != nil {
		return nil, err 
	}

	defer res.Body.Close()

	var rResult struct {
		Hits struct {
			Hits []struct {
				Source models.Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&rResult); err != nil {
		return nil, err 
	}

	products := make([]models.Product, 0)
	for _, hit := range rResult.Hits.Hits {
		products = append(products, hit.Source)
	}

	return products,  nil 

}



func (r *searchRepository) IndexProduct(product *models.Product) error {
    indexName := os.Getenv("ELASTICSEARCH_INDEX")
    if indexName == "" {
        return fmt.Errorf("ELASTICSEARCH_INDEX is not set")
    }

    // Kiểm tra xem index đã tồn tại chưa
    res, err := database.ES.Indices.Exists([]string{indexName})
    if err != nil {
        return fmt.Errorf("failed to check if index exists: %w", err)
    }
    defer res.Body.Close()

    if res.StatusCode == 404 {
        mapping := map[string]interface{}{
            "settings": map[string]interface{}{
                "analysis": map[string]interface{}{
                    "analyzer": map[string]interface{}{
                        "custom_analyzer": map[string]interface{}{
                            "type":      "custom",
                            "tokenizer": "standard",
                            "filter":    []string{"lowercase", "asciifolding"},
                        },
                    },
                },
            },
            "mappings": map[string]interface{}{
                "properties": map[string]interface{}{
                    "name": map[string]interface{}{
                        "type":     "text",
                        "analyzer": "custom_analyzer",
                    },
                    "description": map[string]interface{}{
                        "type":     "text",
                        "analyzer": "custom_analyzer",
                    },
                    "category": map[string]interface{}{
                        "type": "keyword",
                    },
                    "price": map[string]interface{}{
                        "type": "float",
                    },
                    "created_at": map[string]interface{}{
                        "type": "date",
                    },
                    "updated_at": map[string]interface{}{
                        "type": "date",
                    },
                },
            },
        }

        var buf bytes.Buffer
        if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
            return fmt.Errorf("failed to encode index mapping: %w", err)
        }

        createRes, err := database.ES.Indices.Create(indexName, database.ES.Indices.Create.WithBody(&buf))
        if err != nil {
            return fmt.Errorf("failed to create index: %w", err)
        }
        defer createRes.Body.Close()

        if createRes.IsError() {
            return fmt.Errorf("failed to create index: %s", createRes.String())
        }

    }

    // Tiếp tục index sản phẩm
    data, _ := json.Marshal(product)
    _, err = database.ES.Index(
        indexName,
        bytes.NewReader(data),
        database.ES.Index.WithDocumentID(product.ID),
    )

    return err
}


func (r *searchRepository) DeleteProduct(id string) error {
	_, err := database.ES.Delete(
		os.Getenv("ELASTICSEARCH_INDEX"),
		id,
	)

	return err 
}

