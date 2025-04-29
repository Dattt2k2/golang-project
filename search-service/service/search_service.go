package service

import (
	"github.com/Dattt2k2/golang-project/search-service/models" 
	"github.com/Dattt2k2/golang-project/search-service/repository"
)

type SearchService interface {
	BasicSearch(query string) ([]models.Product, error)
	AdvancedSearch(query string, filters map[string]interface{}) ([]models.Product, error)
	IndexProduct(product *models.Product) error
	DeleteProduct(id string) error
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