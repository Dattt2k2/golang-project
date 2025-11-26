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
	UpdateStock(ctx context.Context, id string, quantity int) error
	IncrementSoldCount(ctx context.Context, productID string, quantity int) error
	GetBestSellingProduct(ctx context.Context, limit int) ([]models.Product, error)
	DecrementSoldCount(ctx context.Context, productID string, quantity int) error
	GetProductByCategory(ctx context.Context, category string, skip, limit int64) ([]models.Product, int64, error)
	GetProductStatistics(ctx context.Context) (map[string]int64, error)
	AddProductCategory(ctx context.Context, category string) error
	GetProductCategory(ctx context.Context) ([]models.Category, error)
	DeleteProductCategory(ctx context.Context, categoryID string) error
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
	if product.ID == "" {
		product.ID = uuid.New().String()
	}
	now := time.Now()
	product.Created_at = now
	product.Updated_at = now

	item := map[string]types.AttributeValue{
		"id":          &types.AttributeValueMemberS{Value: product.ID},
		"name":        &types.AttributeValueMemberS{Value: product.Name},
		"description": &types.AttributeValueMemberS{Value: product.Description},
		"price":       &types.AttributeValueMemberN{Value: strconv.FormatFloat(product.Price, 'f', 2, 64)},
		"quantity":    &types.AttributeValueMemberN{Value: strconv.FormatInt(int64(product.Quantity), 10)},
		"category":    &types.AttributeValueMemberS{Value: product.Category},
		"image_path":  &types.AttributeValueMemberSS{Value: product.ImagePath},
		"created_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		"updated_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		"user_id":     &types.AttributeValueMemberS{Value: product.UserID},
		"sold_count":  &types.AttributeValueMemberN{Value: "0"},
		"status":      &types.AttributeValueMemberS{Value: product.Status},
	}

	if len(product.ImagePath) > 0 {
		item["image_path"] = &types.AttributeValueMemberSS{Value: product.ImagePath}
	}

	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	return err
}

func (r *ProductRepositoryImpl) AddProductCategory(ctx context.Context, category string) error {
	now := time.Now()
	categoryItem := map[string]types.AttributeValue{
		"id":         &types.AttributeValueMemberS{Value: uuid.New().String()},
		"name":       &types.AttributeValueMemberS{Value: category},
		"created_at": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
	}

	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("Category"),
		Item: categoryItem,
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

func (r *ProductRepositoryImpl) UpdateStock(ctx context.Context, id string, quantity int) error {
	// Trừ stock khi order thành công (quantity dương = giảm stock)
	logger.Info(fmt.Sprintf("UpdateStock called: productID=%s, quantity=%d, actualValue=%d", id, quantity, -quantity))

	result, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("ADD quantity :qty SET updated_at = :time"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":qty":  &types.AttributeValueMemberN{Value: strconv.Itoa(-quantity)}, // Âm để trừ đi
			":time": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
		ReturnValues: types.ReturnValueAllNew,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to update stock: productID=%s, error=%v", id, err))
		return err
	}

	// Log giá trị mới sau khi update
	if qtyAttr, ok := result.Attributes["quantity"]; ok {
		if qtyN, ok := qtyAttr.(*types.AttributeValueMemberN); ok {
			logger.Info(fmt.Sprintf("Stock updated successfully: productID=%s, new_quantity=%s", id, qtyN.Value))
		}
	}

	return nil
}

func (r *ProductRepositoryImpl) IncrementSoldCount(ctx context.Context, productID string, quantity int) error {
	logger.Info(fmt.Sprintf("IncrementSoldCount called: productID=%s, quantity=%d", productID, quantity))

	result, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
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

	// Log giá trị mới sau khi update
	if soldAttr, ok := result.Attributes["sold_count"]; ok {
		if soldN, ok := soldAttr.(*types.AttributeValueMemberN); ok {
			logger.Info(fmt.Sprintf("Sold count incremented successfully: productID=%s, new_sold_count=%s", productID, soldN.Value))
		}
	}

	return nil
}

func (r *ProductRepositoryImpl) DecrementSoldCount(ctx context.Context, productID string, quantity int) error {
	logger.Info(fmt.Sprintf("DecrementSoldCount called: productID=%s, quantity=%d", productID, quantity))

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

	logger.Info(fmt.Sprintf("Sold count decremented successfully: productID=%s", productID))
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

	if limit > int64(len(products)) {
		limit = int64(len(products))
	}

	if len(products) == 0 {
		return []models.Product{}, total, nil
	}

	return products, total, nil
}

func (r *ProductRepositoryImpl) GetProductStatistics(ctx context.Context) (map[string]int64, error) {
	countResult, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
		Select:    types.SelectCount,
	})
	if err != nil {
		return nil, err
	}

	totalProducts := int64(countResult.Count)

	now := time.Now()
	month := int(now.Month())
	year := now.Year()
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	prevMonth := month - 1
	prevYear := year
	if prevMonth <= 0 {
		prevMonth = 12
		prevYear -= 1
	}

	filterExpr := "#created_at <= :start"
	prevProd, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
		Select: types.SelectCount,
		FilterExpression: aws.String(filterExpr),
		ExpressionAttributeNames: map[string]string{
			"#created_at": "created_at",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{":start": &types.AttributeValueMemberS{Value: start.Format(time.RFC3339)}},
	})
	if err != nil {
		return nil, err
	}

	prevTotalProducts := int64(prevProd.Count)

	stats := map[string]int64{
		"current_total_products": totalProducts,
		"previous_total_products": prevTotalProducts,
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
