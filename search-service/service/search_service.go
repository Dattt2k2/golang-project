package service

import (
	"errors"
	"search-service/models"
	"search-service/repository"
	"strconv"
	"time"
)

type SearchService interface {
	BasicSearch(query string) ([]models.Product, error)
    AdvancedSearch(query string, filters map[string]interface{}, from int, size int, sortBy string, sortOrder string) (models.AdvancedSearchResponse, error)
	IndexProduct(product *models.Product) error
	DeleteProduct(id string) error
	// SyncProductFromProductService() error
}

type searchService struct {
	repo repository.SearchRepository
	S3Service *S3Service
}

func NewSearchService(repo repository.SearchRepository, s3Service *S3Service) SearchService {
	return &searchService{
		repo : repo,
		S3Service: s3Service,
	}
}

func (s *searchService) BasicSearch(query string) ([]models.Product, error) {
	products, err := s.repo.BasicSearch(query)
	if err != nil {
		return nil, err
	}

	for i := range products {
        if len(products[i].ImagePath) > 0 {
            var urls []string
            for _, key := range products[i].ImagePath {
                if key == "" {
                    continue
                }
                url, err := s.GetS3PathIfExist(key, 100*time.Minute)
                if err == nil && url != "" {
                    urls = append(urls, url)
                } else {
                    urls = append(urls, key)
                }
            }
            products[i].ImagePath = urls
        }
    }

	return products, nil
}

func (s *searchService) AdvancedSearch(query string, filters map[string]interface{}, sortBy int, sortOrder int, fromStr string, limitStr string) (models.AdvancedSearchResponse, error) {  // Updated params and return
    from, _ := strconv.Atoi(fromStr)
    size, _ := strconv.Atoi(limitStr)

    products, total, err := s.repo.AdvancedSearch(query, filters, sortBy, sortOrder, fromStr, limitStr)  
    if err != nil {
        return models.AdvancedSearchResponse{}, err
    }

    for i := range products {
        if len(products[i].ImagePath) > 0 {
            var urls []string
            for _, key := range products[i].ImagePath {
                if key == "" {
                    continue
                }
                url, err := s.GetS3PathIfExist(key, 100*time.Minute)
                if err == nil && url != "" {
                    urls = append(urls, url)
                } else {
                    urls = append(urls, key)
                }
            }
            products[i].ImagePath = urls
        }
    }

    havePrev := from > 0
    haveNext := from + len(products) < total
    page := (from / size) + 1

    return models.AdvancedSearchResponse{
        Data:      products,
        Total:     total,
        HavePrev:  havePrev,
        HaveNext:  haveNext,
        Filters:   filters,
        From:      from,
        Limit:     size,
        Page:      page,
        Query:     query,
        SortBy:    strconv.Itoa(sortBy), 
        SortOrder: strconv.Itoa(sortOrder),
    }, nil
}

func (s *searchService) IndexProduct(product *models.Product) error {
	return s.repo.IndexProduct(product)
}

func (s *searchService) DeleteProduct(id string) error {
	return s.repo.DeleteProduct(id)
}

func (s *searchService) GetS3PathIfExist(key string, expiration time.Duration) (string, error) {
	if key == "" {
		return "", errors.New("image key is empty")
	}
	return s.S3Service.GeneratePresignedDownloadURL(key, expiration)
}

// func (s *searchService) SyncProductFromProductService() error {
// 	conn, err := grpc.NewClient("product-service:8089", grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		logger.Error("Failed to connect to product service: %v", logger.ErrField(err))
// 		return err 
// 	}
// 	defer conn.Close()

// 	client := pb.NewProductServiceClient(conn)
// 	resp, err := client.GetAllProducts(context.Background(), &pb.Empty{})
// 	if err != nil {
// 		logger.Error("Failed to get all products from product service: %v", logger.ErrField(err))
// 		return err 
// 	}

// 	for _, p := range resp.Products {
// 		err := s.repo.IndexProduct(&models.Product{
// 			ID:          p.Id,
// 			Name:        p.Name,
// 			Description: p.Description,
// 			Price:       float64(p.Price),
// 			Category:    p.Category,
// 			ImagePath: 	 p.ImagePath,
// 		})
// 		if err != nil {
// 			logger.Err("Failed to index product: %v", err)
// 		}
// 	}

// 	return nil 
// }