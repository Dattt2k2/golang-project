package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"product-service/helper"
	"product-service/kafka"
	"product-service/models"
	"product-service/repository"

	"github.com/google/uuid"
)

type ProductService interface {
	AddProduct(ctx context.Context, product models.Product) error
	EditProduct(ctx context.Context, id string, update map[string]interface{}) error
	DeleteProduct(ctx context.Context, id, userID string) error
	GetProductByID(ctx context.Context, id string) (*models.Product, error)
	// GetProductByName(ctx context.Context, name string) ([]models.Product, error)
	GetAllProducts(ctx context.Context, page, limit int64) ([]models.Product, int64, int, bool, bool, bool, error)
	UpdateProductStock(ctx context.Context, id string, quantity int) error
	IncrementSoldCount(ctx context.Context, productID string, quantity int) error
	GetBestSellingProducts(ctx context.Context, limit int) ([]models.Product, error)
	DecrementSoldCount(ctx context.Context, productID string, quantity int) error
	GetAllProductForIndex(ctx context.Context) ([]models.Product, error)
}

type productServiceImpl struct {
	repo repository.ProductRepository
	S3Service *S3Service
}

func NewProductService(repo repository.ProductRepository, s3Service *S3Service ) ProductService {
	return &productServiceImpl{repo: repo, S3Service : s3Service}
}

func (s *productServiceImpl) AddProduct(ctx context.Context, product models.Product) error {
	product.ID = uuid.New().String()
	product.Created_at = time.Now()
	product.Updated_at = time.Now()
	err := s.repo.Insert(ctx, product)
	if err == nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := helper.InvalidateProductCache(ctx, "products:*"); err != nil {
				log.Printf("Error invalidating product cache: %v", err)
			}
		}()

		go func(p models.Product) {
			_ = kafka.ProduceProductEvent(context.Background(), "created", &p, p.ID)
		}(product)
	}
	
	return err
}

func (s *productServiceImpl) EditProduct(ctx context.Context, id string, update map[string]interface{}) error {
	update["updated_at"] = time.Now()
	err := s.repo.Update(ctx, id, update)
	if err == nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			productKey := fmt.Sprintf("products:%s", id)
			if err := helper.InvalidateProductCache(ctx, productKey); err != nil {
				log.Printf("Error invalidating product cache: %v", err)
			}
		}()
		
		go func(id string) {
			product, err := s.repo.FindByID(context.Background(), id)
			if err == nil && product != nil {
				_ = kafka.ProduceProductEvent(context.Background(), "updated", product, id)
			}
		} (id)
	}
	return err
}

func (s *productServiceImpl) DeleteProduct(ctx context.Context, id, userID string) error {
	err := s.repo.Delete(ctx, id, userID)
	if err == nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			productKey := fmt.Sprintf("product:%s", id)
			if err := helper.InvalidateProductCache(ctx, productKey); err != nil {
				log.Printf("Error invalidating product cache: %v", err)
			}

			if err := helper.InvalidateProductCache(ctx, "products:*"); err != nil {
				log.Printf("Error invalidating product cache: %v", err)
			}
		}()

		go func(id string) {
			_ = kafka.ProduceProductEvent(context.Background(), "deleted", nil, id)
		} (id)
	}
	return err
}

func (s *productServiceImpl) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	cacheKey := fmt.Sprintf("product:%s", id) // Đổi thành "product:" để nhất quán

	var product models.Product
	found, err := helper.GetCachedProductData(ctx, cacheKey, &product)
	if err == nil && found {
		log.Printf("Cache hit for product: %s", id)
		if product.ImagePath != "" {
			url, err := s.GetS3PathIfExist(product.ImagePath, 100*time.Minute)
			if err == nil {
				product.ImagePath = url 
			}
		}
		return &product, nil
	}

	productPtr, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if productPtr != nil {
		go func(p *models.Product) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := helper.CacheProductData(ctx, cacheKey, p, 30*time.Minute); err != nil { 
				log.Printf("Error caching product data: %v", err)
			} else {
				log.Printf("Cached product data with key: %s", cacheKey)
			}
		}(productPtr)
	}
	return productPtr, nil
}

// func (s *productServiceImpl) GetProductByName(ctx context.Context, name string) ([]models.Product, error) {
// 	return s.repo.FindByName(ctx, name)
// }

func (s *productServiceImpl) GetAllProducts(ctx context.Context, page, limit int64) ([]models.Product, int64, int, bool, bool, bool, error) {
	cachedResult, found, err := helper.GetAllProductsFromCache(ctx, page, limit)

	if err == nil && found && cachedResult != nil {
		return cachedResult.Products, cachedResult.Total, cachedResult.Pages, cachedResult.HasNext, cachedResult.HasPrev, true, nil
	}

	skip := (page - 1) * limit
	products, total, err := s.repo.FindAll(ctx, skip, limit)
	if err != nil {
		return nil, 0, 0, false, false, false, err
	}

	for i := range products {
		if products[i].ImagePath != "" {
			url, err := s.GetS3PathIfExist(products[i].ImagePath, 100*time.Minute)
			if err == nil {
				products[i].ImagePath = url 
			}
		}
	}

	pages := int((total + limit - 1) / limit)
	hasNext := page < int64(pages)
	hasPrev := page > 1

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := helper.CacheAllProducts(ctx, page, limit, products, total, pages, hasNext, hasPrev); err != nil {
			log.Printf("Error caching all products: %v", err)
		} else {
			log.Printf("Cached all products for page=%d, limit=%d", page, limit)
		}
	}()
	return products, total, pages, hasNext, hasPrev, false, nil
}

func (s *productServiceImpl) UpdateProductStock(ctx context.Context, id string, quantity int) error {
	err := s.repo.UpdateStock(ctx, id, quantity)
	if err == nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			productKey := fmt.Sprintf("products:%s", id)
			if err := helper.InvalidateProductCache(ctx, productKey); err != nil {
				log.Printf("Error invalidating product cache: %v", err)
			}
		}()
	}
	return err
}

func (s *productServiceImpl) IncrementSoldCount(ctx context.Context, productID string, quantity int) error {
	err := s.repo.IncrementSoldCount(ctx, productID, quantity)
	if err == nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			productKey := fmt.Sprintf("products:%s", productID)
			if err := helper.InvalidateProductCache(ctx, productKey); err != nil {
				log.Printf("Error invalidating product cache: %v", err)
			}
			if err := helper.InvalidateProductCache(ctx, "bestselling:*"); err != nil {
				log.Printf("Error invalidating best selling product cache: %v", err)
			}
		}()
	}
	return err
}

func (s *productServiceImpl) GetBestSellingProducts(ctx context.Context, limit int) ([]models.Product, error) {
	if limit <= 0 {
		limit = 10
	}

	cacheKey := fmt.Sprintf("bestselling:limit=%d", limit)
	var products []models.Product
	found, err := helper.GetCachedProductData(ctx, cacheKey, &products)
	if err == nil && found {
		log.Printf("Cache hit for best selling products: limit=%d", limit)
		return products, nil
	}

	products, err = s.repo.GetBestSellingProduct(ctx, limit)
	if err != nil {
		return nil, err
	}

	go func(prods []models.Product) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := helper.CacheProductData(ctx, cacheKey, prods, 30*time.Minute); err != nil { // Tăng TTL lên 30 phút
			log.Printf("Error caching best selling products: %v", err)
		} else {
			log.Printf("Cached best selling products with key: %s", cacheKey)
		}
	}(products)
	return products, nil
}

func (s *productServiceImpl) DecrementSoldCount(ctx context.Context, productID string, quantity int) error {

	err := s.repo.DecrementSoldCount(ctx, productID, quantity)
	if err == nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			productKey := fmt.Sprintf("products:%s", productID)
			if err := helper.InvalidateProductCache(ctx, productKey); err != nil {
				log.Printf("Error invalidating product cache: %v", err)
			}
			if err := helper.InvalidateProductCache(ctx, "bestselling:*"); err != nil {
				log.Printf("Error invalidating best selling product cache: %v", err)
			}
		}()
	}
	return err
}



func (s *productServiceImpl) GetAllProductForIndex(ctx context.Context) ([]models.Product, error) {
	products, _, err := s.repo.FindAll(ctx, 0, 0)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *productServiceImpl) GetS3PathIfExist(key string, expiration time.Duration) (string, error) {
	if key == "" {
		return "", errors.New("image key is empty")
	}
	return s.S3Service.GeneratePresignedDownloadURL(key, expiration)
}