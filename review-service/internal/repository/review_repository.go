package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"review-service/internal/models"
	logger "review-service/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type ReviewRepository interface {
	Create(ctx context.Context, review models.Review) error
	GetByProductID(ctx context.Context, productID string, limit int, lastKey string) ([]models.Review, string, error)
	GetByID(ctx context.Context, id string) (*models.Review, error)
	AddtoSumPending(ctx context.Context, pending models.SumReviewPending) error
}

type ReviewRepositoryImpl struct {
	client *dynamodb.Client
	table string 
	sumTable string
}

func NewReviewRepository(client *dynamodb.Client, table, sumTable string) ReviewRepository {
	return &ReviewRepositoryImpl{
		client:  client,
		table:   table,
		sumTable: sumTable,
	}
}

func (r *ReviewRepositoryImpl) Create(ctx context.Context, review models.Review) error {
	if review.ID == "" {
		review.ID = uuid.New().String()
	}

	now := time.Now().UTC()
	review.CreatedAt = now

	item := map[string]types.AttributeValue{
		"product_id": &types.AttributeValueMemberS{Value: review.ProductID},
		"review_id":  &types.AttributeValueMemberS{Value: review.ID},
		"user_id":    &types.AttributeValueMemberS{Value: review.UserID},
		"rating":     &types.AttributeValueMemberN{Value: strconv.Itoa(review.Rating)},
		"title":      &types.AttributeValueMemberS{Value: review.Title},
		"body":       &types.AttributeValueMemberS{Value: review.Body},
		"created_at": &types.AttributeValueMemberS{Value: review.CreatedAt.Format(time.RFC3339)},
	}

	 _, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item: item,
	 })

	return err 
}

func (r *ReviewRepositoryImpl) GetByProductID(ctx context.Context, productID string, limit int, lastKey string) ([]models.Review, string, error) {
	input := &dynamodb.QueryInput{
        TableName:              aws.String(r.table),
        KeyConditionExpression: aws.String("product_id = :pid"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":pid": &types.AttributeValueMemberS{Value: productID},
        },
    }

    if limit > 0 {
        input.Limit = aws.Int32(int32(limit))
    }

    if lastKey != "" {
        startKey, err := decodeLastKey(lastKey)
        if err != nil {
            logger.Error("invalid lastKey: " + err.Error())
            return nil, "", err
        }
        input.ExclusiveStartKey = startKey
    }

    result, err := r.client.Query(ctx, input)
    if err != nil {
        logger.Error("Failed to get review")
        return nil, "", err
    }

    if result.Items == nil {
        return []models.Review{}, "", nil
    }

    var reviews []models.Review
    if err := attributevalue.UnmarshalListOfMaps(result.Items, &reviews); err != nil {
        return nil, "", err
    }

    nextKey := ""
    if len(result.LastEvaluatedKey) != 0 {
        enc, err := encodeLastKey(result.LastEvaluatedKey)
        if err != nil {
            return nil, "", err
        }
        nextKey = enc
    }

    return reviews, nextKey, nil
}

func decodeLastKey(encoded string) (map[string]types.AttributeValue, error) {
    raw, err := base64.StdEncoding.DecodeString(encoded)
    if err != nil {
        return nil, err
    }
    var m map[string]interface{}
    if err := json.Unmarshal(raw, &m); err != nil {
        return nil, err
    }
    av, err := attributevalue.MarshalMap(m)
    if err != nil {
        return nil, err
    }
    return av, nil
}

// helper: convert LastEvaluatedKey -> base64(json)
func encodeLastKey(key map[string]types.AttributeValue) (string, error) {
    var tmp map[string]interface{}
    if err := attributevalue.UnmarshalMap(key, &tmp); err != nil {
        return "", err
    }
    raw, err := json.Marshal(tmp)
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(raw), nil
}

func (r *ReviewRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Review, error) {
	return nil, errors.New("not implemented")
}

func (r *ReviewRepositoryImpl) AddtoSumPending(ctx context.Context, pending models.SumReviewPending) error {
	sumTable := r.sumTable 
	if sumTable == "" {
		return errors.New("sum table not configured")
	}
	if pending.ReviewID == "" {
		pending.ReviewID = uuid.New().String()
	}

	now := time.Now()
	pending.LastTimeSum = now

	ttl := time.Now().Add(24 * time.Hour).Unix()

	item := map[string]types.AttributeValue{
		"product_id":    &types.AttributeValueMemberS{Value: pending.ProductID},
		"review_id":     &types.AttributeValueMemberS{Value: pending.ReviewID},
		"last_time_sum": &types.AttributeValueMemberS{Value: pending.LastTimeSum.Format(time.RFC3339)},
		"rating":        &types.AttributeValueMemberN{Value: strconv.Itoa(pending.Rating)},
		"ttl":           &types.AttributeValueMemberN{Value: strconv.FormatInt(ttl, 10)},
	}

	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(sumTable),
		Item: item,
		ConditionExpression: aws.String("attribute_not_exists(review_id)"),
	})

	if err != nil {
		logger.Error("Failed to add to sum pending: " + err.Error())
		return err 
	}
	return nil 
}