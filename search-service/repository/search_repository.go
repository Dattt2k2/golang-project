package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"search-service/database"
	"search-service/models"
)

type SearchRepository interface {
	BasicSearch(query string) ([]models.Product, error)
	AdvancedSearch(query string, filters map[string]interface{}, sortBy int, sortOrder int, from string, limit string) ([]models.Product, int, error)
	IndexProduct(product *models.Product) error
	DeleteProduct(id string) error
}

type searchRepository struct{}

func NewSearchRepository() SearchRepository {
	return &searchRepository{}
}

func (r *searchRepository) BasicSearch(query string) ([]models.Product, error) {
	var buf bytes.Buffer
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
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

func (r *searchRepository) AdvancedSearch(query string, filters map[string]interface{}, sortBy int, sortOrder int, from string, limit string) ([]models.Product, int, error) {
	var buf bytes.Buffer

	fromInt, _ := strconv.Atoi(from)
	sizeInt, _ := strconv.Atoi(limit)

	boolQuery := map[string]interface{}{}
	if strings.TrimSpace(query) == "" {
		boolQuery["must"] = map[string]interface{}{"match_all": map[string]interface{}{}}
	} else {
		boolQuery["must"] = []interface{}{
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  query,
					"fields": []string{"name", "description", "category"},
				},
			},
		}
	}

	// filters array
	filtersArr := []interface{}{}

	// category filter: use match_phrase for text, term for UUID
	if cat, ok := filters["category"]; ok {
		if s, ok := cat.(string); ok && s != "" {
			filtersArr = append(filtersArr, map[string]interface{}{
				"term": map[string]interface{}{"category": s},
			})
		}
	}

	priceFilter := map[string]interface{}{}
	if pmin, ok := filters["price_min"]; ok {
		priceFilter["gte"] = pmin
	}
	if pmax, ok := filters["price_max"]; ok {
		priceFilter["lte"] = pmax
	}
	if len(priceFilter) > 0 {
		filtersArr = append(filtersArr, map[string]interface{}{
			"nested": map[string]interface{}{
				"path": "variants",
				"query": map[string]interface{}{
					"range": map[string]interface{}{"variants.price": priceFilter},
				},
			},
		})
	}

	// Variant filters: size, color, material
	variantFilters := []interface{}{}
	if size, ok := filters["size"]; ok {
		if s, ok := size.(string); ok && s != "" {
			variantFilters = append(variantFilters, map[string]interface{}{
				"term": map[string]interface{}{"variants.size": s},
			})
		}
	}
	if color, ok := filters["color"]; ok {
		if c, ok := color.(string); ok && c != "" {
			variantFilters = append(variantFilters, map[string]interface{}{
				"term": map[string]interface{}{"variants.color": c},
			})
		}
	}
	if material, ok := filters["material"]; ok {
		if m, ok := material.(string); ok && m != "" {
			variantFilters = append(variantFilters, map[string]interface{}{
				"term": map[string]interface{}{"variants.material": m},
			})
		}
	}
	if len(variantFilters) > 0 {
		filtersArr = append(filtersArr, map[string]interface{}{
			"nested": map[string]interface{}{
				"path": "variants",
				"query": map[string]interface{}{
					"bool": map[string]interface{}{"must": variantFilters},
				},
			},
		})
	}

	// rating range
	if rmin, ok := filters["rating_min"]; ok {
		rng := map[string]interface{}{"gte": rmin}
		if rmax, ok := filters["rating_max"]; ok {
			rng["lte"] = rmax
		}
		filtersArr = append(filtersArr, map[string]interface{}{"range": map[string]interface{}{"rating": rng}})
	}

	if len(filtersArr) > 0 {
		boolQuery["filter"] = filtersArr
	}

	// map sortOrder int to string
	order := "desc"
	if sortOrder == 1 {
		order = "asc"
	}

	// map sortBy int to ES field
	var sortClause interface{}
	switch sortBy {
	case 1:
		sortClause = map[string]interface{}{"name.keyword": map[string]interface{}{"order": order}}
	case 2:
		// Sort by min price of variants
		sortClause = map[string]interface{}{
			"variants.price": map[string]interface{}{
				"order": order,
				"mode":  "min",
				"nested": map[string]interface{}{
					"path": "variants",
				},
			},
		}
	case 3:
		sortClause = map[string]interface{}{"rating": map[string]interface{}{"order": order}}
	case 4:
		sortClause = map[string]interface{}{"reviews_count": map[string]interface{}{"order": order}}
	case 5:
		sortClause = map[string]interface{}{"sold_count": map[string]interface{}{"order": order}}
	default:
		sortClause = map[string]interface{}{"created_at": map[string]interface{}{"order": order}}
	}

	esQuery := map[string]interface{}{
		"from":  fromInt,
		"size":  sizeInt,
		"sort":  []interface{}{sortClause},
		"query": map[string]interface{}{"bool": boolQuery},
	}

	if err := json.NewEncoder(&buf).Encode(esQuery); err != nil {
		return nil, 0, err
	}

	res, err := database.ES.Search(
		database.ES.Search.WithContext(context.Background()),
		database.ES.Search.WithIndex(os.Getenv("ELASTICSEARCH_INDEX")),
		database.ES.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, 0, err // Return 0 for total on error
	}
	defer res.Body.Close()

	// Updated struct to include total
	var rResult struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source models.Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&rResult); err != nil {
		return nil, 0, err
	}

	products := make([]models.Product, 0, len(rResult.Hits.Hits))
	for _, hit := range rResult.Hits.Hits {
		products = append(products, hit.Source)
	}
	return products, rResult.Hits.Total.Value, nil
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
						"fields": map[string]interface{}{
							"keyword": map[string]interface{}{
								"type":         "keyword",
								"ignore_above": 256,
							},
						},
					},
					"description": map[string]interface{}{
						"type":     "text",
						"analyzer": "custom_analyzer",
					},
					"category": map[string]interface{}{
						"type": "keyword",
					},
					"variants": map[string]interface{}{
						"type": "nested",
						"properties": map[string]interface{}{
							"id": map[string]interface{}{
								"type": "keyword",
							},
							"size": map[string]interface{}{
								"type": "keyword",
							},
							"color": map[string]interface{}{
								"type": "keyword",
							},
							"material": map[string]interface{}{
								"type": "keyword",
							},
							"cost_price": map[string]interface{}{
								"type": "float",
							},
							"price": map[string]interface{}{
								"type": "float",
							},
							"quantity": map[string]interface{}{
								"type": "integer",
							},
							"created_at": map[string]interface{}{
								"type": "date",
							},
						},
					},
					"created_at": map[string]interface{}{
						"type": "date",
					},
					"updated_at": map[string]interface{}{
						"type": "date",
					},
					"image_path": map[string]interface{}{
						"type": "keyword",
					},
					"review_count": map[string]interface{}{
						"type": "integer",
					},
					"reviews_count": map[string]interface{}{
						"type": "integer",
					},
					"rating_count": map[string]interface{}{
						"type": "integer",
					},
					"sold_count": map[string]interface{}{
						"type": "integer",
					},
					"rating": map[string]interface{}{
						"type": "float",
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
