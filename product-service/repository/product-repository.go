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
        "image_path": &types.AttributeValueMemberSS{Value: product.ImagePath},
        "created_at": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
        "updated_at": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
        "user_id": &types.AttributeValueMemberS{Value: product.UserID},
        "sold_count": &types.AttributeValueMemberN{Value: "0"},
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
func (r *ProductRepositoryImpl) Update(ctx context.Context, id string, update map[string]interface{}) error {
	// Build expression attribute names, values and update expression from the update map
	if update == nil {
		update = map[string]interface{}{}
	}

	exprNames := make(map[string]string)
	exprValues := make(map[string]types.AttributeValue)
	clauses := make([]string, 0, len(update)+1)

    if imagePath, ok := update["image_path"]; ok {
        if paths, ok := imagePath.([]string); ok && len(paths) == 0 {
            exprNames["#image_path"] = "image_path"
            exprValues[":image_path"] = &types.AttributeValueMemberSS{Value: []string{}}
            clauses = append(clauses, "#image_path = :image_path")
            delete(update, "image_path")
        }
    }

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
        TableName: aws.String(r.tableName),
        IndexName: aws.String("category-index"),
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