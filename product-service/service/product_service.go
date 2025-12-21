package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"sort"
	"strings"
	"time"

	"product-service/helper"
	"product-service/kafka"
	"product-service/models"
	"product-service/repository"

	"github.com/google/uuid"
)

type ProductService interface {
	AddProduct(ctx context.Context, product models.Product) error
	EditProduct(ctx context.Context, id string, update map[string]interface{}, categoryName string) error
	DeleteProduct(ctx context.Context, id, userID string) error
	GetProductByID(ctx context.Context, id string) (*models.Product, error)
	GetProductForgRPC(ctx context.Context, id string) (*models.Product, error)
	// GetProductByName(ctx context.Context, name string) ([]models.Product, error)
	GetAllProducts(ctx context.Context, page, limit int64) ([]models.Product, int64, int, bool, bool, bool, error)
	UpdateProductStock(ctx context.Context, productID string, variantID string, quantity int) error
	IncrementSoldCount(ctx context.Context, productID string, quantity int) error
	GetBestSellingProducts(ctx context.Context, limit int) ([]models.Product, error)
	DecrementSoldCount(ctx context.Context, productID string, quantity int) error
	GetAllProductForIndex(ctx context.Context) ([]models.Product, error)
	GetProductByUserID(ctx context.Context, userID string, page, limit int64, category, sortBy, sortOrder string) ([]models.Product, int64, int, bool, bool, error)
	GetProductByCategory(ctx context.Context, category string, page, limit int64) ([]models.Product, int64, int, bool, bool, error)
	GetProductStatistics(ctx context.Context, month, year int) (map[string]interface{}, error)
	AddProductCategory(ctx context.Context, category models.Category) error
	GetProductCategory(ctx context.Context) ([]models.Category, error)
	DeleteProductCategory(ctx context.Context, categoryID string) error
	GetCategoryByName(ctx context.Context, name string) (*models.Category, error)
	GetCategoryByCode(ctx context.Context, code string) (*models.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
}

type productServiceImpl struct {
	repo      repository.ProductRepository
	S3Service *S3Service
}

func NewProductService(repo repository.ProductRepository, s3Service *S3Service) ProductService {
	return &productServiceImpl{repo: repo, S3Service: s3Service}
}

func (s *productServiceImpl) generateProductID(ctx context.Context, categoryName string) string {

	category, err := s.GetCategoryByName(ctx, categoryName)
	if err != nil || category == nil {
		category = &models.Category{
			Code: "GEN",
			Name: "General",
		}
	}

	catCode := category.Code
	randomSuffix := s.generateRandomString(5)
	return fmt.Sprintf("SKU-%s-%s", catCode, randomSuffix)
}

func (s *productServiceImpl) generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ""
		}
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

// generateVariantID tạo variant ID từ productID + size + color + attribute
// Cho phép size/color trống với sản phẩm không cần (như đồ điện tử, trang sức)
func (s *productServiceImpl) generateVariantID(productID, size, color, attribute string) string {
	parts := []string{productID}

	// Normalize và thêm size nếu có
	if size != "" {
		normalizedSize := strings.ToUpper(strings.TrimSpace(removeDiacritics(size)))
		normalizedSize = strings.ReplaceAll(normalizedSize, " ", "-")
		normalizedSize = sanitizeForID(normalizedSize)
		if normalizedSize != "" {
			parts = append(parts, normalizedSize)
		}
	}

	// Normalize và thêm color nếu có
	if color != "" {
		normalizedColor := strings.ToUpper(strings.TrimSpace(removeDiacritics(color)))
		normalizedColor = strings.ReplaceAll(normalizedColor, " ", "-")
		normalizedColor = sanitizeForID(normalizedColor)
		if normalizedColor != "" {
			parts = append(parts, normalizedColor)
		}
	}

	// Normalize và thêm attribute nếu có (dùng cho trang sức, điện tử có nhiều phiên bản)
	if attribute != "" {
		normalizedAttr := strings.ToUpper(strings.TrimSpace(removeDiacritics(attribute)))
		normalizedAttr = strings.ReplaceAll(normalizedAttr, " ", "-")
		normalizedAttr = strings.ReplaceAll(normalizedAttr, ",", ".") // Chuyển dấu phẩy thành dấu chấm
		normalizedAttr = sanitizeForID(normalizedAttr)
		if normalizedAttr != "" {
			parts = append(parts, normalizedAttr)
		}
	}

	// Nếu không có gì, dùng STD
	if len(parts) == 1 {
		parts = append(parts, "STD")
	}

	return strings.Join(parts, "-")
}

// sanitizeForID loại bỏ ký tự không hợp lệ trong ID, chỉ giữ chữ cái, số, dấu chấm và gạch nối
func sanitizeForID(s string) string {
	var result strings.Builder
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '.' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// removeDiacritics bỏ dấu tiếng Việt
func removeDiacritics(s string) string {
	var result strings.Builder
	for _, r := range s {
		switch r {
		case 'à', 'á', 'ạ', 'ả', 'ã', 'â', 'ầ', 'ấ', 'ậ', 'ẩ', 'ẫ', 'ă', 'ằ', 'ắ', 'ặ', 'ẳ', 'ẵ':
			result.WriteRune('a')
		case 'À', 'Á', 'Ạ', 'Ả', 'Ã', 'Â', 'Ầ', 'Ấ', 'Ậ', 'Ẩ', 'Ẫ', 'Ă', 'Ằ', 'Ắ', 'Ặ', 'Ẳ', 'Ẵ':
			result.WriteRune('A')
		case 'è', 'é', 'ẹ', 'ẻ', 'ẽ', 'ê', 'ề', 'ế', 'ệ', 'ể', 'ễ':
			result.WriteRune('e')
		case 'È', 'É', 'Ẹ', 'Ẻ', 'Ẽ', 'Ê', 'Ề', 'Ế', 'Ệ', 'Ể', 'Ễ':
			result.WriteRune('E')
		case 'ì', 'í', 'ị', 'ỉ', 'ĩ':
			result.WriteRune('i')
		case 'Ì', 'Í', 'Ị', 'Ỉ', 'Ĩ':
			result.WriteRune('I')
		case 'ò', 'ó', 'ọ', 'ỏ', 'õ', 'ô', 'ồ', 'ố', 'ộ', 'ổ', 'ỗ', 'ơ', 'ờ', 'ớ', 'ợ', 'ở', 'ỡ':
			result.WriteRune('o')
		case 'Ò', 'Ó', 'Ọ', 'Ỏ', 'Õ', 'Ô', 'Ồ', 'Ố', 'Ộ', 'Ổ', 'Ỗ', 'Ơ', 'Ờ', 'Ớ', 'Ợ', 'Ở', 'Ỡ':
			result.WriteRune('O')
		case 'ù', 'ú', 'ụ', 'ủ', 'ũ', 'ư', 'ừ', 'ứ', 'ự', 'ử', 'ữ':
			result.WriteRune('u')
		case 'Ù', 'Ú', 'Ụ', 'Ủ', 'Ũ', 'Ư', 'Ừ', 'Ứ', 'Ự', 'Ử', 'Ữ':
			result.WriteRune('U')
		case 'ỳ', 'ý', 'ỵ', 'ỷ', 'ỹ':
			result.WriteRune('y')
		case 'Ỳ', 'Ý', 'Ỵ', 'Ỷ', 'Ỹ':
			result.WriteRune('Y')
		case 'đ':
			result.WriteRune('d')
		case 'Đ':
			result.WriteRune('D')
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (s *productServiceImpl) AddProduct(ctx context.Context, product models.Product) error {
	// Generate ProductID từ category
	product.ID = s.generateProductID(ctx, product.Category)

	// Generate VariantID cho từng variant
	for i := range product.Variants {
		product.Variants[i].ID = s.generateVariantID(product.ID, product.Variants[i].Size, product.Variants[i].Color, product.Variants[i].Attribute)
		product.Variants[i].CreatedAt = time.Now()
	}

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

func (s *productServiceImpl) EditProduct(ctx context.Context, id string, update map[string]interface{}, categoryName string) error {
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
		}(id)
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
		}(id)
	}
	return err
}

func (s *productServiceImpl) GetProductForgRPC(ctx context.Context, id string) (*models.Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *productServiceImpl) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	// cacheKey := fmt.Sprintf("product:%s", id) // Đổi thành "product:" để nhất quán

	// var product models.Product
	// found, err := helper.GetCachedProductData(ctx, cacheKey, &product)
	// if err == nil && found {
	// 	log.Printf("Cache hit for product: %s", id)
	// 	if len(product.ImagePath) > 0 {
	// 		var urls []string
	// 		for _, key := range product.ImagePath {
	// 			if key == "" {
	// 				continue
	// 			}
	// 			url, err := s.GetS3PathIfExist(key, 100*time.Minute)
	// 			if err == nil && url != "" {
	// 				urls = append(urls, url)
	// 			} else {
	// 				// fallback to original key if presign fails
	// 				urls = append(urls, key)
	// 			}
	// 		}
	// 		product.ImagePath = urls
	// 	}
	// 	return &product, nil
	// }

	productPtr, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if productPtr != nil && len(productPtr.ImagePath) > 0 {
		urls := make([]string, 0, len(productPtr.ImagePath))
		for _, key := range productPtr.ImagePath {
			if key == "" {
				continue
			}
			url, err := s.GetS3PathIfExist(key, 100*time.Minute)
			if err == nil && url != "" {
				urls = append(urls, url)
			} else {
				// fallback to original key
				urls = append(urls, key)
			}
		}
		productPtr.ImagePath = urls
	}

	// if productPtr != nil {
	// 	go func(p *models.Product) {
	// 		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 		defer cancel()
	// 		if err := helper.CacheProductData(ctx, cacheKey, p, 30*time.Minute); err != nil {
	// 			log.Printf("Error caching product data: %v", err)
	// 		} else {
	// 			log.Printf("Cached product data with key: %s", cacheKey)
	// 		}
	// 	}(productPtr)
	// }
	return productPtr, nil
}

// func (s *productServiceImpl) GetProductByName(ctx context.Context, name string) ([]models.Product, error) {
// 	return s.repo.FindByName(ctx, name)
// }

// sortProducts sorts products based on sortBy field and sortOrder (asc/desc)
func (s *productServiceImpl) sortProducts(products []models.Product, sortBy, sortOrder string) {
	sort.Slice(products, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = products[i].Name < products[j].Name
		case "price":
			// Compare by first variant's price if available
			priceI := int64(0)
			priceJ := int64(0)
			if len(products[i].Variants) > 0 {
				priceI = int64(products[i].Variants[0].Price)
			}
			if len(products[j].Variants) > 0 {
				priceJ = int64(products[j].Variants[0].Price)
			}
			less = priceI < priceJ
		case "sold_count":
			less = products[i].SoldCount < products[j].SoldCount
		case "rating":
			less = products[i].Rating < products[j].Rating
		case "created_at":
			less = products[i].Created_at.Before(products[j].Created_at)
		default:
			// Default to created_at
			less = products[i].Created_at.Before(products[j].Created_at)
		}

		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}

func (s *productServiceImpl) GetAllProducts(ctx context.Context, page, limit int64) ([]models.Product, int64, int, bool, bool, bool, error) {
	// basic validation / defaults
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// Try to get from cache first
	cachedResult, found, err := helper.GetAllProductsFromCache(ctx, page, limit)
	if err != nil {
		log.Printf("Error getting cached products: %v", err)
	}
	if found && cachedResult != nil {
		log.Printf("Cache hit for products page=%d, limit=%d", page, limit)

		// Convert image keys to presigned URLs for cached data
		for i := range cachedResult.Products {
			if len(cachedResult.Products[i].ImagePath) > 0 {
				var urls []string
				for _, key := range cachedResult.Products[i].ImagePath {
					if key == "" {
						continue
					}
					// Check if already a presigned URL (contains X-Amz-Algorithm)
					if strings.Contains(key, "X-Amz-Algorithm") || strings.HasPrefix(key, "http") {
						log.Printf("Skipping presign for already signed URL: %s", key[:50])
						urls = append(urls, key)
						continue
					}
					url, err := s.GetS3PathIfExist(key, 100*time.Minute)
					if err == nil && url != "" {
						urls = append(urls, url)
					} else {
						urls = append(urls, key)
					}
				}
				cachedResult.Products[i].ImagePath = urls
			}
		}

		return cachedResult.Products, cachedResult.Total, cachedResult.Pages, cachedResult.HasNext, cachedResult.HasPrev, true, nil
	}

	// Cache miss - fetch from repository
	skip := (page - 1) * limit
	products, total, err := s.repo.FindAll(ctx, skip, limit)
	if err != nil {
		return nil, 0, 0, false, false, false, err
	}

	// Calculate pages correctly: ceil(total/limit)
	// Use float64 to ensure proper ceiling division, then convert back
	var pages int
	if total == 0 {
		pages = 0
	} else {
		pages = int(math.Ceil(float64(total) / float64(limit)))
	}
	hasNext := page < int64(pages)
	hasPrev := page > 1

	// Cache the result with S3 keys (NOT presigned URLs) để tránh cache URLs đã hết hạn
	go func(prods []models.Product, tot int64, pgs int, hn, hp bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := helper.CacheAllProducts(ctx, page, limit, prods, tot, pgs, hn, hp); err != nil {
			log.Printf("Error caching all products: %v", err)
		} else {
			log.Printf("Cached all products for page=%d, limit=%d", page, limit)
		}
	}(products, total, pages, hasNext, hasPrev)

	// Convert image keys to presigned URLs AFTER caching
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

	return products, total, pages, hasNext, hasPrev, false, nil
}

func (s *productServiceImpl) UpdateProductStock(ctx context.Context, productID string, variantID string, quantity int) error {
	// Invalidate cache TRƯỚC khi update để tránh race condition
	productKey := fmt.Sprintf("products:%s", productID)
	if err := helper.InvalidateProductCache(ctx, productKey); err != nil {
		log.Printf("Warning: Error invalidating product cache before update: %v", err)
	}
	// Invalidate GetAllProducts cache
	if err := helper.InvalidateProductCache(ctx, "products:page=*"); err != nil {
		log.Printf("Warning: Error invalidating products list cache: %v", err)
	}

	err := s.repo.UpdateStock(ctx, productID, variantID, quantity)
	if err != nil {
		return err
	}

	// Invalidate cache sau khi update để đảm bảo
	if err := helper.InvalidateProductCache(ctx, productKey); err != nil {
		log.Printf("Warning: Error invalidating product cache after update: %v", err)
	}
	if err := helper.InvalidateProductCache(ctx, "products:page=*"); err != nil {
		log.Printf("Warning: Error invalidating products list cache after update: %v", err)
	}

	return nil
}

func (s *productServiceImpl) IncrementSoldCount(ctx context.Context, productID string, quantity int) error {
	// Invalidate cache TRƯỚC khi update
	productKey := fmt.Sprintf("products:%s", productID)
	if err := helper.InvalidateProductCache(ctx, productKey); err != nil {
		log.Printf("Warning: Error invalidating product cache before increment: %v", err)
	}
	if err := helper.InvalidateProductCache(ctx, "bestselling:*"); err != nil {
		log.Printf("Warning: Error invalidating bestselling cache: %v", err)
	}
	// Invalidate GetAllProducts cache vì sold_count thay đổi ảnh hưởng sort
	if err := helper.InvalidateProductCache(ctx, "products:page=*"); err != nil {
		log.Printf("Warning: Error invalidating products list cache: %v", err)
	}

	err := s.repo.IncrementSoldCount(ctx, productID, quantity)
	if err != nil {
		return err
	}

	// Invalidate cache sau khi update để đảm bảo
	if err := helper.InvalidateProductCache(ctx, productKey); err != nil {
		log.Printf("Warning: Error invalidating product cache after increment: %v", err)
	}
	if err := helper.InvalidateProductCache(ctx, "bestselling:*"); err != nil {
		log.Printf("Warning: Error invalidating bestselling cache after increment: %v", err)
	}
	if err := helper.InvalidateProductCache(ctx, "products:page=*"); err != nil {
		log.Printf("Warning: Error invalidating products list cache after increment: %v", err)
	}

	return nil
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
	products, _, err := s.repo.FindAll(ctx, 0, 1000)
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

func (s *productServiceImpl) GetProductByUserID(ctx context.Context, userID string, page, limit int64, category, sortBy, sortOrder string) ([]models.Product, int64, int, bool, bool, error) {
	// Fetch ALL products for user (no skip/limit yet) when filtering is needed
	var products []models.Product
	var total int64
	var err error

	// Lookup category name from code if category filter provided
	var categoryName string
	if category != "" {
		cat, err := s.GetCategoryByCode(ctx, category)
		if err == nil && cat != nil {
			categoryName = cat.Name
		} else {
			// If not found by code, try using the value as-is
			categoryName = category
		}
	}

	if categoryName != "" {
		// Need to fetch all to filter properly
		products, total, err = s.repo.FindByUserID(ctx, userID, 0, 9999)
		if err != nil {
			return nil, 0, 0, false, false, err
		}

		// Filter by category name (products store category name, not code)
		var filtered []models.Product
		for _, p := range products {
			if p.Category == categoryName {
				filtered = append(filtered, p)
			}
		}
		products = filtered
		total = int64(len(products))
	} else {
		// No filter, can use pagination at DB level
		skip := (page - 1) * limit
		products, total, err = s.repo.FindByUserID(ctx, userID, skip, limit)
		if err != nil {
			return nil, 0, 0, false, false, err
		}
	}

	// Apply sorting
	s.sortProducts(products, sortBy, sortOrder)

	// Apply pagination AFTER filtering and sorting
	if categoryName != "" {
		skip := (page - 1) * limit
		start := skip
		if start > int64(len(products)) {
			return []models.Product{}, total, 0, false, false, nil
		}
		end := start + limit
		if end > int64(len(products)) {
			end = int64(len(products))
		}
		products = products[start:end]
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
					// fallback to original key if presign fails
					urls = append(urls, key)
				}
			}
			products[i].ImagePath = urls
		}
	}

	// Calculate pages correctly: ceil(total/limit)
	// Use float64 to ensure proper ceiling division, then convert back
	var pages int
	if total == 0 {
		pages = 0
	} else {
		pages = int(math.Ceil(float64(total) / float64(limit)))
	}
	hasNext := page < int64(pages)
	hasPrev := page > 1

	return products, total, pages, hasNext, hasPrev, nil
}

func (s *productServiceImpl) GetProductByCategory(ctx context.Context, category string, page, limit int64) ([]models.Product, int64, int, bool, bool, error) {
	skip := (page - 1) * limit

	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	cacheKey := fmt.Sprintf("products:category=%s:page=%d:limit%d", category, page, limit)

	var cached struct {
		Products []models.Product `json:"products"`
		Total    int64            `json:"total"`
		Pages    int              `json:"pages"`
		HasNext  bool             `json:"has_next"`
		HasPrev  bool             `json:"has_prev"`
	}

	found, err := helper.GetCachedProductData(ctx, cacheKey, &cached)
	if err == nil && found {
		return cached.Products, cached.Total, cached.Pages, cached.HasNext, cached.HasPrev, nil
	}

	products, total, err := s.repo.GetProductByCategory(ctx, category, skip, limit)
	if err != nil {
		return nil, 0, 0, false, false, err
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
					// fallback to original key if presign fails
					urls = append(urls, key)
				}
			}
			products[i].ImagePath = urls
		}
	}

	// Calculate pages correctly: ceil(total/limit)
	// Use float64 to ensure proper ceiling division, then convert back
	var pages int
	if total == 0 {
		pages = 0
	} else {
		pages = int(math.Ceil(float64(total) / float64(limit)))
	}
	hasNext := page < int64(pages)
	hasPrev := page > 1

	go func(cat string, p int64, l int64, prods []models.Product, tot int64, pgs int, hn, hp bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cachePayload := map[string]interface{}{
			"products": prods,
			"total":    tot,
			"pages":    pgs,
			"has_next": hn,
			"has_prev": hp,
		}
		if err := helper.CacheProductData(ctx, cacheKey, cachePayload, 30*time.Minute); err != nil {
			log.Printf("Failed to cache product data for category %s: %v", cat, err)
		}
	}(category, page, limit, products, total, pages, hasNext, hasPrev)

	return products, total, pages, hasNext, hasPrev, nil
}

func (s *productServiceImpl) GetProductStatistics(ctx context.Context, month, year int) (map[string]interface{}, error) {

	top, err := s.repo.GetBestSellingProduct(ctx, 5)
	if err != nil {
		return nil, err
	}

	stats, err := s.repo.GetProductStatistics(ctx, month, year)
	if err != nil {
		return nil, err
	}

	totalProd := stats["current_total_products"]
	prevTotalProd := stats["previous_total_products"]

	var growth float64
	if prevTotalProd == 0 {
		if totalProd == 0 {
			growth = 0
		} else {
			growth = 100.0
		}
	} else {
		growth = (float64(totalProd) - float64(prevTotalProd)) / float64(prevTotalProd) * 100
		growth = math.Round(growth*100) / 100
	}

	topSold := make([]map[string]interface{}, 0, len(top))
	for _, p := range top {
		if p.ID == "" {
			continue
		}
		// Lấy giá từ variant đầu tiên (nếu có)
		var price float64
		if len(p.Variants) > 0 {
			price = p.Variants[0].Price
		}
		topSold = append(topSold, map[string]interface{}{
			"product_id": p.ID,
			"name":       p.Name,
			"sold_count": p.SoldCount,
			"price":      price,
		})
	}

	// Add debug log so you can inspect values
	log.Printf("product-stats: total=%d prev=%d growth=%.2f topCount=%d", totalProd, prevTotalProd, growth, len(topSold))

	resp := map[string]interface{}{
		"top_selling_products":    topSold,
		"total_products":          totalProd,
		"previous_total_products": prevTotalProd,
		"growth_percentage":       growth,
	}

	return resp, nil
}

func (s *productServiceImpl) AddProductCategory(ctx context.Context, category models.Category) error {
	category.ID = uuid.New().String()
	category.CreatedAt = time.Now()
	return s.repo.AddProductCategory(ctx, category.Name, category.Code)
}

func (s *productServiceImpl) GetProductCategory(ctx context.Context) ([]models.Category, error) {
	return s.repo.GetProductCategory(ctx)
}

func (s *productServiceImpl) DeleteProductCategory(ctx context.Context, categoryID string) error {
	category, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return err
	}
	if category == nil {
		return fmt.Errorf("category not found")
	}

	count, err := s.repo.CountProductsByCategoryName(ctx, category.Name)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot delete category '%s': %d product(s) reference it", category.Name, count)
	}

	return s.repo.DeleteProductCategory(ctx, categoryID)
}

func (s *productServiceImpl) GetCategoryByName(ctx context.Context, name string) (*models.Category, error) {
	return s.repo.GetCategoryByName(ctx, name)
}

func (s *productServiceImpl) GetCategoryByCode(ctx context.Context, code string) (*models.Category, error) {
	return s.repo.GetCategoryByCode(ctx, code)
}

func (s *productServiceImpl) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	return s.repo.GetCategoryByID(ctx, id)
}
