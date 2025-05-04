package service

import (
	"context"

	pb "github.com/Dattt2k2/golang-project/module/gRPC-Product/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Dattt2k2/golang-project/search-service/log"
	"github.com/Dattt2k2/golang-project/search-service/models"
	"github.com/Dattt2k2/golang-project/search-service/repository"
)

type SearchService interface {
	BasicSearch(query string) ([]models.Product, error)
	AdvancedSearch(query string, filters map[string]interface{}) ([]models.Product, error)
	IndexProduct(product *models.Product) error
	DeleteProduct(id string) error
	SyncProductFromProductService() error
}

type searchService struct {
	repo repository.SearchRepository
}

func NewSearchService(repo repository.SearchRepository) SearchService {
	return &searchService{
		repo : repo,
	}
}

func (s *searchService) BasicSearch(query string) ([]models.Product, error) {
	return s.repo.BasicSearch(query)
}

func (s *searchService) AdvancedSearch(query string, filters map[string]interface{}) ([]models.Product, error) {
	return s.repo.AdvancedSearch(query, filters)
}

func (s *searchService) IndexProduct(product *models.Product) error {
	return s.repo.IndexProduct(product)
}

func (s *searchService) DeleteProduct(id string) error {
	return s.repo.DeleteProduct(id)
}

func (s *searchService) SyncProductFromProductService() error {
	conn, err := grpc.NewClient("product-service:8089", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to connect to product service: %v", logger.ErrField(err))
		return err 
	}
	defer conn.Close()

	client := pb.NewProductServiceClient(conn)
	resp, err := client.GetAllProducts(context.Background(), &pb.Empty{})
	if err != nil {
		logger.Error("Failed to get all products from product service: %v", logger.ErrField(err))
		return err 
	}

	logger.Info("Syncing products from product service to search service", logger.Int("total_products", len(resp.Products)))

	for _, p := range resp.Products {
		err := s.repo.IndexProduct(&models.Product{
			ID:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       float64(p.Price),
			Category:    p.Category,
			ImageURL: 	 p.ImageUrl,
		})
		if err != nil {
			logger.Err("Failed to index product: %v", err)
		}
	}
	return nil 
}