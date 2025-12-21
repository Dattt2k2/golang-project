package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	logger "product-service/log"
	"product-service/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type ProductRepository interface {
	Insert(ctx context.Context, product models.Product) error
	Update(ctx context.Context, id string, update map[string]interface{}) error
	Delete(ctx context.Context, id, userID string) error
	FindByID(ctx context.Context, id string) (*models.Product, error)
	// FindByName(ctx context.Context, name string) ([]models.Product, error)
	FindAll(ctx context.Context, skip, limit int64) ([]models.Product, int64, error)
	FindByUserID(ctx context.Context, userID string, skip, limit int64) ([]models.Product, int64, error)
	UpdateStock(ctx context.Context, productID string, variantID string, quantity int) error
	IncrementSoldCount(ctx context.Context, productID string, quantity int) error
	GetBestSellingProduct(ctx context.Context, limit int) ([]models.Product, error)
	DecrementSoldCount(ctx context.Context, productID string, quantity int) error
	GetProductByCategory(ctx context.Context, category string, skip, limit int64) ([]models.Product, int64, error)
	GetProductStatistics(ctx context.Context, month, year int) (map[string]int64, error)
	AddProductCategory(ctx context.Context, category string, code string) error
	GetProductCategory(ctx context.Context) ([]models.Category, error)
	DeleteProductCategory(ctx context.Context, categoryID string) error
	GetCategoryByName(ctx context.Context, name string) (*models.Category, error)
	GetCategoryByCode(ctx context.Context, code string) (*models.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
	CountProductsByCategoryName(ctx context.Context, categoryName string) (int64, error)
}

type ProductRepositoryImpl struct {
	client    *dynamodb.Client
	tableName string
}

func NewProductRepository(client *dynamodb.Client, tableName string) ProductRepository {
	return &ProductRepositoryImpl{
		client:    client,
		tableName: tableName,
	}
}

func (r *ProductRepositoryImpl) Insert(ctx context.Context, product models.Product) error {

	now := time.Now()
	product.Created_at = now
	product.Updated_at = now

	// Marshal variants to DynamoDB format
	variantsAttr, err := attributevalue.MarshalList(product.Variants)
	if err != nil {
		return fmt.Errorf("failed to marshal variants: %w", err)
	}

	item := map[string]types.AttributeValue{
		"id":          &types.AttributeValueMemberS{Value: product.ID},
		"name":        &types.AttributeValueMemberS{Value: product.Name},
		"description": &types.AttributeValueMemberS{Value: product.Description},
		"variants":    &types.AttributeValueMemberL{Value: variantsAttr},
		"category":    &types.AttributeValueMemberS{Value: product.Category},
		"created_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		"updated_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		"user_id":     &types.AttributeValueMemberS{Value: product.UserID},
		"sold_count":  &types.AttributeValueMemberN{Value: "0"},
		"status":      &types.AttributeValueMemberS{Value: product.Status},
	}

	// Only add image_path if not empty - DynamoDB doesn't allow empty string sets
	if len(product.ImagePath) > 0 {
		item["image_path"] = &types.AttributeValueMemberSS{Value: product.ImagePath}
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	return err
}

func (r *ProductRepositoryImpl) AddProductCategory(ctx context.Context, category string, code string) error {
	now := time.Now()
	categoryItem := map[string]types.AttributeValue{
		"id":         &types.AttributeValueMemberS{Value: uuid.New().String()},
		"name":       &types.AttributeValueMemberS{Value: category},
		"code":       &types.AttributeValueMemberS{Value: code},
		"created_at": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
	}

	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("Category"),
		Item:      categoryItem,
	})
	return err
}

func (r *ProductRepositoryImpl) GetProductCategory(ctx context.Context) ([]models.Category, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String("Category"),
	})
	if err != nil {
		return nil, err
	}
	var categories []models.Category
	for _, item := range result.Items {
		var category models.Category
		err = attributevalue.UnmarshalMap(item, &category)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *ProductRepositoryImpl) DeleteProductCategory(ctx context.Context, categoryID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String("Category"),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: categoryID},
		},
	})
	return err
}

func (r *ProductRepositoryImpl) Update(ctx context.Context, id string, update map[string]interface{}) error {
	if update == nil {
		update = map[string]interface{}{}
	}

	exprNames := make(map[string]string)
	exprValues := make(map[string]types.AttributeValue)
	clauses := make([]string, 0, len(update)+1)

	// Handle empty image_path explicitly before the general loop
	if imagePath, ok := update["image_path"]; ok {
		if paths, ok := imagePath.([]string); ok && len(paths) == 0 {
			exprNames["#image_path"] = "image_path"
			exprValues[":image_path"] = &types.AttributeValueMemberSS{Value: []string{}}
			clauses = append(clauses, "#image_path = :image_path")
			delete(update, "image_path")
		}
	}

	// Remove updated_at from update map if it exists (we'll add it ourselves)
	delete(update, "updated_at")

	for k, v := range update {
		if k == "image_path" {
			continue
		}

		nameKey := "#" + k
		valKey := ":" + k
		exprNames[nameKey] = k

		av, err := attributevalue.Marshal(v)
		if err != nil {
			return err
		}
		exprValues[valKey] = av
		clauses = append(clauses, fmt.Sprintf("%s = %s", nameKey, valKey))
	}

	// Always update updated_at
	exprNames["#updated_at"] = "updated_at"
	updatedAtVal, err := attributevalue.Marshal(time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}
	exprValues[":updated_at"] = updatedAtVal
	clauses = append(clauses, "#updated_at = :updated_at")

	updateExpr := "SET " + strings.Join(clauses, ", ")

	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
		UpdateExpression:          aws.String(updateExpr),
	})
	return err
}

func (r *ProductRepositoryImpl) Delete(ctx context.Context, id, userID string) error {
	product, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if product.UserID != userID {
		return fmt.Errorf("unauthorized: user does not own the product")
	}

	_, err = r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

// func (r *ProductRepositoryImpl) FindByID(ctx context.Context, id string) (*models.Product, error) {
// 	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
// 		TableName: aws.String(r.tableName),
// 		Key: map[string]types.AttributeValue{
// 			"id": &types.AttributeValueMemberS{Value: id},
// 		},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	if result.Item == nil {
// 		return nil, fmt.Errorf("product not found")
// 	}

//		var product models.Product
//		err = attributevalue.UnmarshalMap(result.Item, &product)
//		if err != nil {
//			return nil, err
//		}
//		return &product, nil
//	}
func (r *ProductRepositoryImpl) FindByID(ctx context.Context, id string) (*models.Product, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})

	if err != nil {
		logger.Err("DynamoDB GetItem error", err)
		return nil, err
	}

	if result.Item == nil {
		return nil, fmt.Errorf("product not found")
	}

	prod, err := decodeProduct(result.Item)
	if err != nil {
		logger.Err("Failed to decode product", err)
		return nil, err
	}

	return &prod, nil
}

// func (r *productRepositoryImpl) FindByName(ctx context.Context, name string) ([]models.Product, error) {
// 	var products []models.Product
// 	filter := bson.M{"name":bson.M{"$regex": name, "$options": "i"}}
// 	cursor, err := r.collection.Find(ctx, filter)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer cursor.Close(ctx)
// 	for cursor.Next(ctx) {
// 		var product models.Product
// 		if err := cursor.Decode(&product); err != nil {
// 			return nil, err
// 		}
// 		products = append(products, product)
// 	}
// 	return products, nil
// }

func (r *ProductRepositoryImpl) FindAll(ctx context.Context, skip, limit int64) ([]models.Product, int64, error) {
	countResult, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
		Select:    types.SelectCount,
	})
	if err != nil {
		return nil, 0, err
	}

	total := int64(countResult.Count)

	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
		Limit:     aws.Int32(int32(limit)),
	}
	var products []models.Product
	var scannedCount int64 = 0

	paginator := dynamodb.NewScanPaginator(r.client, scanInput)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, 0, err
		}

		for _, item := range page.Items {
			if scannedCount < skip {
				scannedCount++
				continue
			}

			if int64(len(products)) >= limit {
				break
			}

			product, err := decodeProduct(item)
			if err != nil {
				logger.Err("unmarshal product", err)
				continue
			}

			products = append(products, product)
			scannedCount++
		}
	}
	return products, total, nil
}

func (r *ProductRepositoryImpl) UpdateStock(ctx context.Context, productID string, variantID string, quantity int) error {
	// Lấy thông tin product hiện tại

	product, err := r.FindByID(ctx, productID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to find product: productID=%s, error=%v", productID, err))
		return err
	}

	// Tìm và cập nhật variant
	variantFound := false
	for i := range product.Variants {
		if product.Variants[i].ID == variantID {
			// Trừ stock khi order thành công (quantity dương nghĩa là trừ đi)
			product.Variants[i].Quantity -= quantity

			if product.Variants[i].Quantity < 0 {
				return fmt.Errorf("insufficient stock for variant %s", variantID)
			}

			variantFound = true
			break
		}
	}

	if !variantFound {
		return fmt.Errorf("variant not found: variantID=%s", variantID)
	}

	// Marshal variants mới
	variantsAttr, err := attributevalue.MarshalList(product.Variants)
	if err != nil {
		return fmt.Errorf("failed to marshal variants: %w", err)
	}

	// Cập nhật vào DynamoDB
	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: productID},
		},
		UpdateExpression: aws.String("SET variants = :variants, updated_at = :time"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":variants": &types.AttributeValueMemberL{Value: variantsAttr},
			":time":     &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
		ReturnValues: types.ReturnValueAllNew,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to update stock: productID=%s, error=%v", productID, err))
		return err
	}

	return nil
}

func (r *ProductRepositoryImpl) IncrementSoldCount(ctx context.Context, productID string, quantity int) error {

	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: productID},
		},
		UpdateExpression: aws.String("ADD sold_count :qty SET updated_at = :time"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":qty":  &types.AttributeValueMemberN{Value: strconv.Itoa(quantity)},
			":time": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
		ReturnValues: types.ReturnValueAllNew,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to increment sold count: productID=%s, error=%v", productID, err))
		return err
	}

	return nil
}

func (r *ProductRepositoryImpl) DecrementSoldCount(ctx context.Context, productID string, quantity int) error {

	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: productID},
		},
		UpdateExpression: aws.String("ADD sold_count :qty SET updated_at = :time"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":qty":  &types.AttributeValueMemberN{Value: strconv.Itoa(-quantity)}, // Âm để trừ
			":time": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to decrement sold count: productID=%s, error=%v", productID, err))
		return err
	}

	return nil
}

func (r *ProductRepositoryImpl) GetBestSellingProduct(ctx context.Context, limit int) ([]models.Product, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	})
	if err != nil {
		return nil, err
	}

	var products []models.Product
	for _, item := range result.Items {
		var product models.Product
		err = attributevalue.UnmarshalMap(item, &product)
		if err != nil {
			continue
		}
		products = append(products, product)
	}

	// Sort by sold_count in descending order
	for i := 0; i < len(products)-1; i++ {
		for j := 0; j < len(products)-i-1; j++ {
			if products[j].SoldCount < products[j+1].SoldCount {
				products[j], products[j+1] = products[j+1], products[j]
			}
		}
	}

	// Return only the requested limit
	if limit > len(products) {
		limit = len(products)
	}

	return products[:limit], nil
}

func (r *ProductRepositoryImpl) FindByUserID(ctx context.Context, userID string, skip, limit int64) ([]models.Product, int64, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("user_id-index"),
		KeyConditionExpression: aws.String("#user_id = :uid"),
		ExpressionAttributeNames: map[string]string{
			"#user_id": "user_id",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: userID},
		},
	}

	var products []models.Product
	var total int64 = 0
	var skipped int64 = 0

	paginator := dynamodb.NewQueryPaginator(r.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, 0, err
		}
		for _, item := range page.Items {
			total++
			if skipped < skip {
				skipped++
				continue
			}
			if int64(len(products)) >= limit {
				break
			}
			product, err := decodeProduct(item)
			if err != nil {
				logger.Err("unmarshal product", err)
				continue
			}
			products = append(products, product)
		}
		if int64(len(products)) >= limit {
			break
		}
	}

	return products, total, nil
}

func (r *ProductRepositoryImpl) GetProductByCategory(ctx context.Context, category string, skip, limit int64) ([]models.Product, int64, error) {
	getItemInput := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("category-index"),
		KeyConditionExpression: aws.String("#category = :cat"),
		ExpressionAttributeNames: map[string]string{
			"#category": "category",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":cat": &types.AttributeValueMemberS{Value: category},
		},
	}

	var products []models.Product
	var total int64 = 0

	paginator := dynamodb.NewQueryPaginator(r.client, getItemInput)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, 0, err
		}
		for _, item := range page.Items {
			product, err := decodeProduct(item)
			if err != nil {
				logger.Err("unmarshal product", err)
				continue
			}
			products = append(products, product)
			total++
		}
	}

	for i := 0; i < len(products)-1; i++ {
		for j := 0; j < len(products)-i-1; j++ {
			if products[j].SoldCount < products[j+1].SoldCount {
				products[j], products[j+1] = products[j+1], products[j]
			}
		}
	}

	if len(products) == 0 {
		return []models.Product{}, total, nil
	}

	// Apply pagination
	start := skip
	if start > int64(len(products)) {
		return []models.Product{}, total, nil
	}

	end := start + limit
	if end > int64(len(products)) {
		end = int64(len(products))
	}

	return products[start:end], total, nil
}

func (r *ProductRepositoryImpl) GetCategoryByName(ctx context.Context, name string) (*models.Category, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String("Category"),
		FilterExpression: aws.String("#name = :nameVal"),
		ExpressionAttributeNames: map[string]string{
			"#name": "name",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":nameVal": &types.AttributeValueMemberS{Value: name},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var category models.Category
	err = attributevalue.UnmarshalMap(result.Items[0], &category)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *ProductRepositoryImpl) GetCategoryByCode(ctx context.Context, code string) (*models.Category, error) {
	logger.Logger.Infof("[GetCategoryByCode] Scanning for code: '%s'", code)

	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:        aws.String("Category"),
			FilterExpression: aws.String("code = :codeVal"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":codeVal": &types.AttributeValueMemberS{Value: code},
			},
			Limit: aws.Int32(25),
		}

		if lastEvaluatedKey != nil {
			input.ExclusiveStartKey = lastEvaluatedKey
		}

		result, err := r.client.Scan(ctx, input)
		if err != nil {
			logger.Logger.Errorf("[GetCategoryByCode] Error scanning: %v", err)
			return nil, err
		}

		logger.Logger.Infof("[GetCategoryByCode] Scanned page, found %d items", len(result.Items))

		if len(result.Items) > 0 {
			var category models.Category
			err = attributevalue.UnmarshalMap(result.Items[0], &category)
			if err != nil {
				logger.Logger.Errorf("[GetCategoryByCode] Error unmarshaling: %v", err)
				return nil, err
			}
			logger.Logger.Infof("[GetCategoryByCode] Found category: code='%s', name='%s'", category.Code, category.Name)
			return &category, nil
		}

		if result.LastEvaluatedKey == nil {
			logger.Logger.Infof("[GetCategoryByCode] No more pages, category not found: '%s'", code)
			break
		}

		lastEvaluatedKey = result.LastEvaluatedKey
	}

	return nil, nil
}

func (r *ProductRepositoryImpl) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("Category"),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var category models.Category
	err = attributevalue.UnmarshalMap(result.Item, &category)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *ProductRepositoryImpl) CountProductsByCategoryName(ctx context.Context, categoryName string) (int64, error) {
	// Use category-index to count items with this category
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("category-index"),
		KeyConditionExpression: aws.String("#category = :cat"),
		ExpressionAttributeNames: map[string]string{
			"#category": "category",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":cat": &types.AttributeValueMemberS{Value: categoryName},
		},
		Select: types.SelectCount,
	}

	paginator := dynamodb.NewQueryPaginator(r.client, input)
	var total int64 = 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return 0, err
		}
		total += int64(page.Count)
	}

	return total, nil
}

func (r *ProductRepositoryImpl) GetProductStatistics(ctx context.Context, month, year int) (map[string]int64, error) {
	countResult, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
		Select:    types.SelectCount,
	})
	if err != nil {
		return nil, err
	}
	totalProducts := int64(countResult.Count)

	now := time.Now()

	if month < 0 || month > 12 {
		return nil, fmt.Errorf("invalid month: %d", month)
	}
	if year < 0 {
		return nil, fmt.Errorf("invalid year: %d", year)
	}

	if month == 0 && year == 0 {
		month = int(now.Month())
		year = now.Year()
	}
	if month != 0 && year == 0 {
		year = now.Year()
	}

	var currentStart, currentEnd, prevStart, prevEnd time.Time
	if month > 0 {
		currentStart = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		currentEnd = currentStart.AddDate(0, 1, 0)
		prevStart = currentStart.AddDate(0, -1, 0)
		prevEnd = currentStart
	} else {
		currentStart = time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		currentEnd = currentStart.AddDate(1, 0, 0)
		prevStart = currentStart.AddDate(-1, 0, 0)
		prevEnd = currentStart
	}

	countBetween := func(start, end time.Time) (int64, error) {
		filterExpr := "#created_at >= :start AND #created_at < :end"
		out, err := r.client.Scan(ctx, &dynamodb.ScanInput{
			TableName:        aws.String(r.tableName),
			Select:           types.SelectCount,
			FilterExpression: aws.String(filterExpr),
			ExpressionAttributeNames: map[string]string{
				"#created_at": "created_at",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":start": &types.AttributeValueMemberS{Value: start.Format(time.RFC3339)},
				":end":   &types.AttributeValueMemberS{Value: end.Format(time.RFC3339)},
			},
		})
		if err != nil {
			return 0, err
		}
		return int64(out.Count), nil
	}

	currentTotal, err := countBetween(currentStart, currentEnd)
	if err != nil {
		return nil, err
	}

	prevTotal, err := countBetween(prevStart, prevEnd)
	if err != nil {
		return nil, err
	}

	stats := map[string]int64{
		"current_total_products":  currentTotal,
		"previous_total_products": prevTotal,
		"total_products":          totalProducts,
	}

	return stats, nil
}

func decodeProduct(item map[string]types.AttributeValue) (models.Product, error) {
	var p models.Product

	itemCopy := make(map[string]types.AttributeValue, len(item))
	for k, v := range item {
		if k == "image_path" {
			continue
		}
		itemCopy[k] = v
	}

	if err := attributevalue.UnmarshalMap(itemCopy, &p); err != nil {
		return p, err
	}

	if av, ok := item["image_path"]; ok && av != nil {
		switch v := av.(type) {
		case *types.AttributeValueMemberSS:
			p.ImagePath = v.Value
		case *types.AttributeValueMemberS:
			if v.Value == "" {
				p.ImagePath = []string{}
			} else if strings.Contains(v.Value, ",") {
				parts := strings.Split(v.Value, ",")
				for i := range parts {
					parts[i] = strings.TrimSpace(parts[i])
				}
				p.ImagePath = parts
			} else {
				p.ImagePath = []string{v.Value}
			}
		case *types.AttributeValueMemberL:
			out := make([]string, 0, len(v.Value))
			for _, elem := range v.Value {
				if s, ok := elem.(*types.AttributeValueMemberS); ok {
					out = append(out, s.Value)
				}
			}
			p.ImagePath = out
		default:
			p.ImagePath = []string{}
		}
	} else {
		p.ImagePath = []string{}
	}

	return p, nil
}
