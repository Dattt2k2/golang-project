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
		"id": &types.AttributeValueMemberS{Value: product.ID},
		"name": &types.AttributeValueMemberS{Value: product.Name},
		"description": &types.AttributeValueMemberS{Value: product.Description},
		"price": &types.AttributeValueMemberN{Value: strconv.FormatFloat(product.Price, 'f', 2, 64)},
		"quantity": &types.AttributeValueMemberN{Value: strconv.FormatInt(int64(product.Quantity), 10)},
		"category": &types.AttributeValueMemberS{Value: product.Category},
		"image_path": &types.AttributeValueMemberS{Value: product.ImagePath},
		"created_at": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		"updated_at": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		"user_id": &types.AttributeValueMemberS{Value: product.UserID},
		"sold_count": &types.AttributeValueMemberN{Value: "0"},
	}

	if product.ImagePath != "" {
		item["image_path"] = &types.AttributeValueMemberS{Value: product.ImagePath}
	}

	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	return err
}
func (r *ProductRepositoryImpl) Update(ctx context.Context, id string, update map[string]interface{}) error {
	// Build expression attribute names, values and update expression from the update map
	if update == nil {
		update = map[string]interface{}{}
	}

	exprNames := make(map[string]string)
	exprValues := make(map[string]types.AttributeValue)
	clauses := make([]string, 0, len(update)+1)

	for k, v := range update {
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

	// always update updated_at
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

// 	var product models.Product 
// 	err = attributevalue.UnmarshalMap(result.Item, &product)
// 	if err != nil {
// 		return nil, err 
// 	}
// 	return &product, nil
// }
func (r *ProductRepositoryImpl) FindByID(ctx context.Context, id string) (*models.Product, error) {
    // Debug log
    logger.Info("Finding product by ID in DynamoDB", logger.Str("id", id))
    
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
        logger.Info("Product not found in DynamoDB", logger.Str("id", id))
        return nil, fmt.Errorf("product not found")
    }
    
    var product models.Product
    err = attributevalue.UnmarshalMap(result.Item, &product)
    if err != nil {
        logger.Err("Failed to unmarshal DynamoDB item", err)
        return nil, err
    }
    
    logger.Info("Successfully found product", 
        logger.Str("id", id),
        logger.Str("name", product.Name))
    
    return &product, nil
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

func (r *ProductRepositoryImpl) FindAll(ctx context.Context, skip, limit int64) ([]models.Product, int64, error){
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

			var product models.Product
			err = attributevalue.UnmarshalMap(item, &product)
			if err != nil {
				return nil, 0, err 
			}

			products = append(products, product)
			scannedCount++
		}
	}
	return products, total, nil
}

func (r *ProductRepositoryImpl) UpdateStock(ctx context.Context, id string, quantity int) error {
    _, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
        TableName: aws.String(r.tableName),
        Key: map[string]types.AttributeValue{
            "id": &types.AttributeValueMemberS{Value: id},
        },
        UpdateExpression: aws.String("ADD quantity :qty SET updated_at = :time"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":qty":  &types.AttributeValueMemberN{Value: strconv.Itoa(quantity)},
            ":time": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
        },
    })
    return err
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
    })
    return err
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

func (r *ProductRepositoryImpl) DecrementSoldCount(ctx context.Context, productID string, quantity int) error {
    _, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
        TableName: aws.String(r.tableName),
        Key: map[string]types.AttributeValue{
            "id": &types.AttributeValueMemberS{Value: productID},
        },
        UpdateExpression: aws.String("ADD sold_count :qty SET updated_at = :time"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":qty":  &types.AttributeValueMemberN{Value: strconv.Itoa(-quantity)},
            ":time": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
        },
    })
    return err
}

func (r *ProductRepositoryImpl) FindByUserID(ctx context.Context, userID string, skip, limit int64) ([]models.Product, int64, error) {
	input := &dynamodb.QueryInput{
        TableName: aws.String(r.tableName),
        IndexName: aws.String("user_id-index"),
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
            var p models.Product
            if err := attributevalue.UnmarshalMap(item, &p); err != nil {
                // skip malformed item but continue
                logger.Err("unmarshal product", err)
                continue
            }
            products = append(products, p)
        }
        if int64(len(products)) >= limit {
            break
        }
    }

    return products, total, nil
}