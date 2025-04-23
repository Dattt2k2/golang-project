package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Dattt2k2/golang-project/product-service/database"
	"github.com/Dattt2k2/golang-project/product-service/models"
	"github.com/Dattt2k2/golang-project/product-service/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductService interface {
	AddProduct(ctx context.Context, product models.Product) error
	EditProduct(ctx context.Context, id primitive.ObjectID, update bson.M) error
	DeleteProduct(ctx context.Context, id, userID primitive.ObjectID) error
	GetProductByID(ctx context.Context, id primitive.ObjectID) (*models.Product, error)
	GetProductByName(ctx context.Context, name string) ([]models.Product, error)
	GetAllProducts(ctx context.Context, page, limit int64) ([]models.Product,int64, int, bool, bool, error)
	UpdateProductStock(ctx context.Context, id primitive.ObjectID, quantity int) error
	IncrementSoldCount(ctx context.Context, productID string, quantity int) error
	GetBestSellingProducts(ctx context.Context, limit int) ([]models.Product, error)
	DecrementSoldCount(ctx context.Context, productID string, quantity int) error
}

type productServiceImpl struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productServiceImpl{repo: repo}
}

func (s *productServiceImpl) AddProduct(ctx context.Context, product models.Product) error {
	product.ID = primitive.NewObjectID()
	product.Created_at = time.Now()
	product.Updated_at = time.Now()
	return s.repo.Insert(ctx, product)
}

func (s *productServiceImpl) EditProduct(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()
	return s.repo.Update(ctx, id, update)
}

func (s *productServiceImpl) DeleteProduct(ctx context.Context, id, userID primitive.ObjectID) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *productServiceImpl) GetProductByID(ctx context.Context, id primitive.ObjectID) (*models.Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *productServiceImpl) GetProductByName(ctx context.Context, name string) ([]models.Product, error) {
	return s.repo.FindByName(ctx, name)
}

func (s *productServiceImpl) GetAllProducts(ctx context.Context, page, limit int64) ([]models.Product,int64, int, bool, bool, error) {
	cacheKey := fmt.Sprintf("products:page=%d&limit=%d", page, limit)

	if database.RedisClient != nil {
		cachedData, err := database.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedResult struct {
				Products []models.Product `json:"products"`
				Total    int64           `json:"total"`
				Pages    int             `json:"pages"`
				HasNext  bool            `json:"has_next"`
				HasPrev  bool            `json:"has_prev"`
			}

			if err := json.Unmarshal([]byte(cachedData), &cachedResult); err == nil {
				log.Printf("cache hit for products: page=%d, limit=%d", page, limit )
				return cachedResult.Products, cachedResult.Total ,cachedResult.Pages,
				cachedResult.HasNext, cachedResult.HasPrev, nil 
			}
		}
	}

	
	skip := int64(page -1) *limit 
	products, total, err := s.repo.FindAll(ctx, skip, int64(limit))
	if err != nil {
		return nil, 0, 0, false, false, err 
	}
	pages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < int64(pages) 
	hasPrev := page > 1
	return products, total, pages, hasNext, hasPrev, nil
}

func (s *productServiceImpl) UpdateProductStock(ctx context.Context, id primitive.ObjectID, quantity int) error {
	return s.repo.UpdateStock(ctx, id, quantity)
}

func (s *productServiceImpl) IncrementSoldCount(ctx context.Context, productID string, quantity int) error {
	productIDObj, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return errors.New("invalid product ID")
	}

	return s.repo.IncrementSoldCount(ctx, productIDObj, quantity)
}

func (s *productServiceImpl) GetBestSellingProducts(ctx context.Context, limit int) ([]models.Product, error) {
	if limit <= 0 {
		limit = 10 
	}

	return s.repo.GetBestSellingProduct(ctx, limit)
}

func (s *productServiceImpl) DecrementSoldCount(ctx context.Context, productID string, quantity int) error {
	productIDObj, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return errors.New("invalid product ID")
	}

	return s.repo.DecrementSoldCount(ctx, productIDObj, quantity)
	
}