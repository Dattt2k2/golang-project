// package repositories

// import (
// 	"context"
// 	"log"

// 	"github.com/Dattt2k2/golang-project/order-service/models"
// 	"github.com/Dattt2k2/golang-project/auth-service/database"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// type OrderRepository struct{
// 	collection *mongo.Collection
// }

// func NewOrderRepository(collection *mongo.Collection) *OrderRepository {
// 	collection = database.OpenCollection(database.Client, "orders")
// 	return &OrderRepository{collection: collection}
// }

// func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) (primitive.ObjectID, error){
// 	result, err := r.collection.InsertOne(ctx, order)
// 	if err != nil{
// 		log.Printf("Failed to create ordrer: %v", err)
// 		return primitive.NilObjectID, err
// 	}

// 	return result.InsertedID.(primitive.ObjectID), nil
// }

// func (r *OrderRepository) UpdateOrder(ctx context.Context, order models.Order) error {
// 	filter := bson.M{"_id": order.ID}
// 	update := bson.M{"$set": order}

// 	_, err := r.collection.UpdateOne(ctx, filter, update)
// 	if err != nil{
// 		log.Printf("Failed to update order %s: %v",order.ID.Hex(), err)
// 		return err
// 	}
// 	return nil
// }

// func (r *OrderRepository) FindOrders(ctx context.Context, page, limit int) ([]models.Order, int64, error) {
// 	skip := (page - 1) * limit

// 	total, err := r.collection.CountDocuments(ctx, bson.M{})
// 	if err != nil{
// 		log.Printf("Failed to count orders: %v", err)
// 		return nil, 0, err
// 	}

// 	if total == 0 {
// 		return []models.Order{}, 0, nil
// 	}

// 	finOptions := options.Find().
// 		SetSkip(int64(skip)).
// 		SetLimit(int64(limit)).
// 		SetSort(bson.D{{Key:"created_at",Value:  -1}})

// 	cursor, err := r.collection.Find(ctx, bson.M{}, finOptions)
// 	if err != nil{
// 		log.Printf("Failed to find orders: %v", err)
// 		return nil, 0, err
// 	}

// 	defer cursor.Close(ctx)

// 	var orders []models.Order
// 	if err := cursor.All(ctx, &orders); err != nil{
// 		log.Printf("Failed to decode orders: %v", err)
// 		return nil, 0, err
// 	}
// 	return orders, total, nil
// }

// func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID primitive.ObjectID) ([]models.OrderItem, error) {
// 	var items []models.OrderItem

// 	pipeline := mongo.Pipeline{
//         bson.D{{Key: "$match", Value: bson.M{"_id": orderID}}},
//         bson.D{{Key: "$unwind", Value: bson.M{"path": "$items"}}},
//         bson.D{{Key: "$lookup", Value: bson.M{
//             "product_id": "$items.product_id",
//             "quantity": "$items.quantity",
//             "price": "$items.price",
//             "name": "$items.name",
//         }}},
// 	}

// 	cursor, err := r.collection.Aggregate(ctx, pipeline)
// 	if err != nil{
// 		log.Printf("Failed to aggregate order items: %v", err)
// 		return nil, err
// 	}
// 	defer cursor.Close(ctx)

// 	if err := cursor.All(ctx, &items); err != nil{
// 		log.Printf("Failed to decode order items: %v", err)
// 		return nil, err
// 	}

// 	return items, nil
// }

// func (r *OrderRepository) FindOrderById(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]models.Order, int64, error) {
// 	skip := (page -1) * limit

// 	filter := bson.M{"user_id": userID}
// 	total, err := r.collection.CountDocuments(ctx, filter)
// 	if err  != nil{
// 		log.Printf("Failed to count orders: %v", err)
// 		return nil, 0, err
// 	}

// 	if total == 0 {
// 		return []models.Order{}, 0, nil
// 	}

// 	findOptions := options.Find().
// 		SetSkip(int64(skip)).
// 		SetLimit(int64(limit)).
// 		SetSort(bson.D{{Key: "created_at", Value: -1}})

// 	cursor, err := r.collection.Find(ctx, filter, findOptions)
// 	if err != nil{
// 		log.Printf("Failed to find orders: %v", err)
// 		return nil, 0, err
// 	}

// 	defer cursor.Close(ctx)

// 	var orders []models.Order
// 	if err := cursor.All(ctx, &orders); err != nil{
// 		log.Printf("Failed to decode orders: %v", err)
// 		return nil, 0, err
// 	}

// 	return orders, total, nil
// }

package repositories

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/Dattt2k2/golang-project/order-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(collection *mongo.Collection) *OrderRepository {
	return &OrderRepository{
		collection: collection,
	}
}

// CreateOrder inserts a new order into the database
func (r *OrderRepository) CreateOrder(ctx context.Context, order models.Order) (*models.Order, error) {
	_, err := r.collection.InsertOne(ctx, order)
	if err != nil {
		log.Printf("Error inserting order: %v", err)
		return nil, err
	}
	return &order, nil
}

// FindOrders retrieves all orders with pagination
func (r *OrderRepository) FindOrders(ctx context.Context, page, limit int) ([]models.Order, int64, error) {
	log.Printf("Repository - FindOrders: page=%d, limit=%d", page, limit)

	// Log tổng số documents trong collection
	totalInCollection, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("Failed to count total documents: %v", err)
		return nil, 0, err
	}
	log.Printf("Repository - Total documents in collection: %d", totalInCollection)

	// Không có đơn hàng nào
	if totalInCollection == 0 {
		return []models.Order{}, 0, nil
	}

	skip := (page - 1) * limit

	// Lấy đơn hàng theo trang
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		log.Printf("Failed to find documents: %v", err)
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		log.Printf("Failed to decode documents: %v", err)
		return nil, 0, err
	}

	log.Printf("Repository - Found %d orders", len(orders))

	return orders, totalInCollection, nil
}

// FindOrdersByUserID retrieves orders for a specific user with pagination
func (r *OrderRepository) FindOrdersByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]models.Order, int64, error) {
	log.Printf("Repository - FindOrdersByUserID: userID=%s, page=%d, limit=%d", userID.Hex(), page, limit)

	filter := bson.M{"user_id": userID}

	// Đếm tổng số đơn hàng của user
	userOrderCount, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("Failed to count user orders: %v", err)
		return nil, 0, err
	}
	log.Printf("Repository - Documents matching userID %s: %d", userID.Hex(), userOrderCount)

	// Không có đơn hàng nào
	if userOrderCount == 0 {
		return []models.Order{}, 0, nil
	}

	skip := (page - 1) * limit

	// Lấy đơn hàng theo trang
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("Failed to find user orders: %v", err)
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		log.Printf("Failed to decode user orders: %v", err)
		return nil, 0, err
	}

	log.Printf("Repository - Found %d orders for user %s", len(orders), userID.Hex())

	return orders, userOrderCount, nil
}

// GetOrderItems retrieves items for a specific order
func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID primitive.ObjectID) ([]models.OrderItem, error) {
	log.Printf("Repository - GetOrderItems: orderID=%s", orderID.Hex())

	var order models.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Order not found: %s", orderID.Hex())
			return []models.OrderItem{}, nil
		}
		log.Printf("Error finding order: %v", err)
		return nil, err
	}

	return order.Items, nil
}

// FindOrdersWithFilter retrieves orders with custom filters and pagination
func (r *OrderRepository) FindOrdersWithFilter(ctx context.Context, filter bson.M, page, limit int) ([]models.Order, int64, error) {
	log.Printf("Repository - FindOrdersWithFilter: filter=%v, page=%d, limit=%d", filter, page, limit)

	// Đếm tổng số đơn hàng phù hợp với bộ lọc
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("Failed to count filtered orders: %v", err)
		return nil, 0, err
	}
	log.Printf("Repository - Documents matching filter: %d", total)

	// Không có đơn hàng nào
	if total == 0 {
		return []models.Order{}, 0, nil
	}

	skip := (page - 1) * limit

	// Lấy đơn hàng theo trang
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("Failed to find filtered orders: %v", err)
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		log.Printf("Failed to decode filtered orders: %v", err)
		return nil, 0, err
	}

	log.Printf("Repository - Found %d orders with filter", len(orders))

	return orders, total, nil
}

// GetOrderItemDetails retrieves detailed information for items in an order
// This replicates the aggregation pipeline from the commented code
func (r *OrderRepository) GetOrderItemDetails(ctx context.Context, orderID primitive.ObjectID) ([]models.OrderItem, error) {
	log.Printf("Repository - GetOrderItemDetails: orderID=%s", orderID.Hex())

	// Dùng pipeline để lấy chi tiết sản phẩm
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"_id": orderID}}},
		bson.D{{Key: "$unwind", Value: bson.M{"path": "$items"}}},
		bson.D{{Key: "$project", Value: bson.M{
			"product_id":  "$items.product_id",
			"quantity":    "$items.quantity",
			"price":       "$items.price",
			"name":        "$items.name",
			"image_url":   "$items.image_url",
			"description": "$items.description",
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Error fetching order items: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []models.OrderItem
	if err := cursor.All(ctx, &items); err != nil {
		log.Printf("Failed to decode order items: %v", err)
		return nil, err
	}

	log.Printf("Repository - Found %d item details for order %s", len(items), orderID.Hex())

	return items, nil
}

// CalculateOrderPages calculates pagination info based on total orders and limit
func CalculateOrderPages(total int64, limit int) int {
	return int(math.Ceil(float64(total) / float64(limit)))
}

// UpdateOrderStatus updates the status of an order
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID primitive.ObjectID, status string) error {
	log.Printf("Repository - UpdateOrderStatus: orderID=%s, status=%s", orderID.Hex(), status)

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": orderID}, update)
	if err != nil {
		log.Printf("Failed to update order status: %v", err)
		return err
	}

	return nil
}

// FindOrderByID retrieves a specific order by ID
func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID primitive.ObjectID) (*models.Order, error) {
	log.Printf("Repository - FindOrderByID: orderID=%s", orderID.Hex())

	var order models.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Order not found: %s", orderID.Hex())
			return nil, nil
		}
		log.Printf("Error finding order: %v", err)
		return nil, err
	}

	return &order, nil
}
