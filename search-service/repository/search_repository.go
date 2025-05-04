package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"os"

	"github.com/Dattt2k2/golang-project/search-service/database"
	"github.com/Dattt2k2/golang-project/search-service/models"
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
	data, _ := json.Marshal(product)
	_, err := database.ES.Index(
		os.Getenv("ELASTICSEARCH_INDEX"),
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

