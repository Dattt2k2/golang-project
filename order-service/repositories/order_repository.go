package repositories

import (
	"context"
	"log"

	"github.com/Dattt2k2/golang-project/order-service/models"
	"github.com/Dattt2k2/golang-project/auth-service/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderRepository struct{
	collection *mongo.Collection
}

func NewOrderRepository(collection *mongo.Collection) *OrderRepository {
	collection = database.OpenCollection(database.Client, "orders")
	return &OrderRepository{collection: collection}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) (primitive.ObjectID, error){
	result, err := r.collection.InsertOne(ctx, order)
	if err != nil{
		log.Printf("Failed to create ordrer: %v", err)
		return primitive.NilObjectID, err
	}

	return result.InsertedID.(primitive.ObjectID), nil
}

func (r *OrderRepository) FindOrderById(ctx context.Context, id primitive.ObjectID) (*models.Order, error){
	var order models.Order

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil{
		log.Printf("Failed to find order: %v", err)
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, order models.Order) error {
	filter := bson.M{"_id": order.ID}
	update := bson.M{"$set": order}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil{
		log.Printf("Failed to update order %s: %v",order.ID.Hex(), err)
		return err
	}
	return nil
}

func (r *OrderRepository) FindOrders(ctx context.Context, page, limit int) ([]models.Order, int64, error) {
	skip := (page - 1) * limit

	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil{
		log.Printf("Failed to count orders: %v", err)
		return nil, 0, err
	}

	if total == 0 {
		return []models.Order{}, 0, nil
	}

	finOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key:"created_at",Value:  -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, finOptions)
	if err != nil{
		log.Printf("Failed to find orders: %v", err)
		return nil, 0, err
	}

	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil{
		log.Printf("Failed to decode orders: %v", err)
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID primitive.ObjectID) ([]models.OrderItem, error) {
	var items []models.OrderItem

	pipeline := mongo.Pipeline{
        bson.D{{Key: "$match", Value: bson.M{"_id": orderID}}},
        bson.D{{Key: "$unwind", Value: bson.M{"path": "$items"}}},
        bson.D{{Key: "$lookup", Value: bson.M{
            "product_id": "$items.product_id",
            "quantity": "$items.quantity",
            "price": "$items.price",
            "name": "$items.name",
        }}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil{
		log.Printf("Failed to aggregate order items: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &items); err != nil{
		log.Printf("Failed to decode order items: %v", err)
		return nil, err
	}

	return items, nil
}