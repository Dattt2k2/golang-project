package repository

import (
	"context"
	"time"

	"github.com/Dattt2k2/golang-project/cart-service/database"
	"github.com/Dattt2k2/golang-project/cart-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CartRepository interface {
	AddItem(ctx context.Context, userID primitive.ObjectID, item models.CartItem) error
	FindByUserID(ctx context.Context, userID primitive.ObjectID) (*models.Cart, error)
	RemoveItem(ctx context.Context, userID primitive.ObjectID, productID primitive.ObjectID) (int64, error)
	ClearCart(ctx context.Context, userID primitive.ObjectID) error
	GetAllCarts(ctx context.Context, page, limit int) ([]models.Cart, int64, error)
	GetCartItems(ctx context.Context, userID primitive.ObjectID) ([]models.CartItem, error)
}

type cartRepositoryImpl struct {
	collection *mongo.Collection
}

func NewcartRepository() CartRepository {
	collection := database.OpenCollection(database.Client, "cart")
	return &cartRepositoryImpl{collection: collection}
}

func (r *cartRepositoryImpl) AddItem(ctx context.Context, userID primitive.ObjectID, item models.CartItem) error {
	update := bson.M{
		"$push" : bson.M{
			"items": bson.M{
				"$each": []models.CartItem{item},
				"$position": 0,
			},
		},
		"$set": bson.M{"updated_at": time.Now()},
	}

	opt := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, bson.M{"user_id": userID}, update, opt)
	return err 
}


func (r *cartRepositoryImpl) FindByUserID(ctx context.Context, userID primitive.ObjectID) (*models.Cart, error) {
	var cart models.Cart
	filter := bson.M{"user_id": userID}
	err := r.collection.FindOne(ctx, filter).Decode(&cart)
	if err != nil {
		return nil, err 
	}

	return &cart, nil
}

func (r *cartRepositoryImpl) RemoveItem(ctx context.Context, userID primitive.ObjectID, productID primitive.ObjectID) (int64, error) {
	filter := bson.M{
		"user_id": userID,
		"items": bson.M{
			"$elemMatch": bson.M{
				"product_id": productID,
			},
		},
	}

	update := bson.M{
		"$pull": bson.M{
			"items": bson.M{
				"product_id": productID,
			},
		},
	}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	return result.ModifiedCount, err 
}

func (r *cartRepositoryImpl) ClearCart(ctx context.Context, userID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": bson.M{"items": []models.CartItem{}}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err 
}

func (r *cartRepositoryImpl) GetAllCarts(ctx context.Context, page, limit int) ([]models.Cart, int64, error) {
	var carts []models.Cart
	skip := (page - 1) * limit

	total , err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err 
	}

	if total == 0 {
		return []models.Cart{}, 0, nil
	}

	findOptions := options.Find(). 
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, 0, err 
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &carts); err != nil {
		return nil, 0, err 
	}

	return carts, total, nil
}

func (r *cartRepositoryImpl) GetCartItems(ctx context.Context, userID primitive.ObjectID) ([]models.CartItem, error) {
	var items []models.CartItem
	
	itemCursor, err := r.collection.Aggregate(ctx, mongo.Pipeline{
        bson.D{{Key: "$match", Value: bson.M{"_id": userID}}},
        bson.D{{Key: "$unwind", Value: "$items"}},
        bson.D{{Key: "$project", Value: bson.M{
            "product_id": "$items.product_id",
            "quantity":   "$items.quantity",
            "price":      "$items.price",
            "name":       "$items.name",
            "image_url":  "$items.image_url",
        }}},
    })

	if err != nil {
		return nil, err 
	}

	defer itemCursor.Close(ctx)

	if err := itemCursor.All(ctx, &items); err != nil {
		return nil, err 
	}

	return items, nil
}

