package repositories

import (
	"context"
	"log"
	"time"

	"github.com/Dattt2k2/golang-project/order-service/models"
	"github.com/Dattt2k2/golang-project/product-service/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepo struct {
	collection *mongo.Collection
}


func NewOrderRepo() *OrderRepo{
	collection := database.OpenCollection(database.Client, "orders")
	return &OrderRepo{collection: collection}
}

func (or *OrderRepo) CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	order.Created_at = time.Now()
	order.Updated_at = time.Now()

	_, err := or.collection.InsertOne(ctx, order)
	if err != nil{
		log.Println("Error creating order:", err)
		return nil, err
	}

	return order, nil
} 

func (or *OrderRepo) FindOrderByID(ctx context.Context, id string) (*models.Order, error){
	var order models.Order
	filter := bson.M{"_id":id}

	err := or.collection.FindOne(ctx, filter).Decode(&order)
	if err != nil{
		if err  == mongo.ErrNoDocuments{
			return nil, nil
		}
		log.Println("Error finding order:", err)
		return nil, err
	}
	return &order, nil
}


