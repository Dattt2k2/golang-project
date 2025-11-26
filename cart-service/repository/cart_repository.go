package repository

import (
	"context"
	"errors"
	"time"

	"cart-service/models"

	// "github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type CartRepository interface {
	AddItem(ctx context.Context, userID string, item models.CartItem) error
	FindByUserID(ctx context.Context, userID string) (*models.Cart, error)
	RemoveItem(ctx context.Context, userID string, productID string) (int64, error)
	ClearCart(ctx context.Context, userID string) error
	GetAllCarts(ctx context.Context, page, limit int) ([]models.Cart, int64, error)
	GetCartItems(ctx context.Context, userID string) ([]models.CartItem, error)
	UpdateCartItem(ctx context.Context, userID string, productID string, quantity int) error
}

type cartRepositoryImpl struct {
	client    *dynamodb.Client
	tableName string
}

func NewCartRepository(client *dynamodb.Client, tableName string) CartRepository {
	return &cartRepositoryImpl{client: client, tableName: tableName}
}

func (r *cartRepositoryImpl) AddItem(ctx context.Context, userID string, item models.CartItem) error {
	itemAV, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err 
	}

	cart, err := r.FindByUserID(ctx, userID)
	if err != nil {
		return err 
	}

	if cart != nil {
		for _, cartItem := range cart.Items {
			if cartItem.ProductID == item.ProductID {
				return errors.New("item already exists in cart")
			}
		}
	}

	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
		},

		UpdateExpression: aws.String("SET #items = list_append(if_not_exists(#items, :empty_list), :new_item),updated_at = :updated_at, created_at = if_not_exists(created_at, :created_at)"),
		ExpressionAttributeNames: map[string]string{
			"#items": "items",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":new_item":    &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberM{Value: itemAV}}},
			":empty_list": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
			":updated_at":  &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
			":created_at":  &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	return err 
}

func (r *cartRepositoryImpl) FindByUserID(ctx context.Context, userID string) (*models.Cart, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
		},
	})

	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}
	var cart models.Cart
	err = attributevalue.UnmarshalMap(result.Item, &cart)
	if err != nil {
		return nil, err 
	}

	return &cart, nil
}

// func (r *cartRepositoryImpl) FindByUserID(ctx context.Context, userID primitive.ObjectID) (*models.Cart, error) {
// 	var cart models.Cart
// 	filter := bson.M{"user_id": userID}
// 	err := r.collection.FindOne(ctx, filter).Decode(&cart)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &cart, nil
// }

func (r *cartRepositoryImpl) RemoveItem(ctx context.Context, userID string, productID string) (int64, error) {
	cart, err := r.FindByUserID(ctx, userID)
	if err != nil {
		return 0, err 
	}

	newItems := make([]models.CartItem, 0)
	removed := false 

	for _, item := range cart.Items {
		if item.ProductID == productID {
			removed = true
		} else {
			newItems = append(newItems, item)
		}
	}

	if !removed {
		return 0, nil 
	}

	cart.Items = newItems
	cart.Updated_at = time.Now()

	cartItem, err := attributevalue.MarshalMap(cart)
	if err != nil {
		return 0, err
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      cartItem,
	})
	if err != nil {
		return 0, err 
	}
	return 1, nil
}

func (r *cartRepositoryImpl) ClearCart(ctx context.Context, userID string) error {
	cart, err := r.FindByUserID(ctx, userID)
	if err != nil {
		return err 
	}

	cart.Items = []models.CartItem{}
	cart.Updated_at = time.Now()

	cartItem, err := attributevalue.MarshalMap(cart)
	if err != nil {
		return err 
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      cartItem,
	})
	return err
}

func (r *cartRepositoryImpl) GetAllCarts(ctx context.Context, page, limit int) ([]models.Cart, int64, error) {
	countResult, err := r.client.Scan(ctx, &dynamodb.ScanInput{
        TableName: aws.String(r.tableName),
        Select:    types.SelectCount,
    })
    if err != nil {
        return nil, 0, err
    }
    total := countResult.Count

    scanInput := &dynamodb.ScanInput{
        TableName: aws.String(r.tableName),
        Limit:     aws.Int32(int32(limit)),
    }

    var carts []models.Cart
    var scannedCount int64 = 0
    skip := int64((page - 1) * limit)

    paginator := dynamodb.NewScanPaginator(r.client, scanInput)
    for paginator.HasMorePages() {
        result, err := paginator.NextPage(ctx)
        if err != nil {
            return nil, 0, err
        }

        for _, item := range result.Items {
            if scannedCount < skip {
                scannedCount++
                continue
            }

            if int64(len(carts)) >= int64(limit) {
                break
            }

            var cart models.Cart
            err = attributevalue.UnmarshalMap(item, &cart)
            if err != nil {
                continue
            }

            carts = append(carts, cart)
            scannedCount++
        }

        if int64(len(carts)) >= int64(limit) {
            break
        }
    }

    return carts, int64(total), nil
}

func (r *cartRepositoryImpl) GetCartItems(ctx context.Context, userID string) ([]models.CartItem, error) {
	cart, err := r.FindByUserID(ctx, userID)
    if err != nil {
        return nil, err
    }

    return cart.Items, nil
}

func (r *cartRepositoryImpl) UpdateCartItem(ctx context.Context, userID string, productID string, quantity int) error {
	if quantity == 0 {
		_, err := r.RemoveItem(ctx, userID, productID)
		return err
	}

	cart, err := r.FindByUserID(ctx, userID)
	if err != nil {
		return err 
	}

	if cart == nil {
		return errors.New("cart not found")
	}
	found := false
	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			cart.Items[i].Quantity = quantity
			cart.Updated_at = time.Now()
			found = true
			break
		}
	}

	if !found {
		return errors.New("product not found in cart")
	}

	cartItem, err := attributevalue.MarshalMap(cart)
	if err != nil {
		return err 
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      cartItem,
	})
	return err
}