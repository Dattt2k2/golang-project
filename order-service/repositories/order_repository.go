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
	"math"
	"time"

	"github.com/Dattt2k2/golang-project/cart-service/log"
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
		logger.Err("Error inserting order", err, logger.Str("order_id", order.ID.Hex()))
		return nil, err
	}
	return &order, nil
}

// FindOrders retrieves all orders with pagination
func (r *OrderRepository) FindOrders(ctx context.Context, page, limit int) ([]models.Order, int64, error) {
	logger.Info("Repository - FindOrders", logger.Int("page", page), logger.Int("limit", limit))

	// Log tổng số documents trong collection
	totalInCollection, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		logger.Info("Failed to count total documents", logger.ErrField(err))
		return nil, 0, err
	}
	logger.Logger.Infof("Repository - Total documents in collection")

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
		logger.Err("Failed to find documents", err, logger.Int("page", page), logger.Int("limit", limit))
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		logger.Err("Failed to decode documents", err, logger.Int("page", page), logger.Int("limit", limit))
		return nil, 0, err
	}

	logger.Logger.Infof("Repository - Found %d orders", len(orders))

	return orders, totalInCollection, nil
}

// FindOrdersByUserID retrieves orders for a specific user with pagination
func (r *OrderRepository) FindOrdersByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]models.Order, int64, error) {
	logger.Info("Repository - FindOrdersByUserID", logger.Str("user_id", userID.Hex()), logger.Int("page", page), logger.Int("limit", limit))

	filter := bson.M{"user_id": userID}

	// Đếm tổng số đơn hàng của user
	userOrderCount, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Err("Failed to count user orders", err, logger.Str("user_id", userID.Hex()))
		return nil, 0, err
	}
	logger.Logger.Infof("Repository - Total documents matching userID %s", userID.Hex())

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
		logger.Err("Failed to find user orders", err, logger.Str("user_id", userID.Hex()))
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		logger.Err("Failed to decode user orders", err, logger.Str("user_id", userID.Hex()))
		return nil, 0, err
	}

	logger.Logger.Infof("Repository - Found %d orders for user %s", len(orders), userID.Hex())

	return orders, userOrderCount, nil
}

// GetOrderItems retrieves items for a specific order
func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID primitive.ObjectID) ([]models.OrderItem, error) {
	logger.Info("Repository - GetOrderItems", logger.Str("order_id", orderID.Hex()))

	var order models.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Info("Order not found", logger.Str("order_id", orderID.Hex()))
			return []models.OrderItem{}, nil
		}
		logger.Err("Error finding order", err, logger.Str("order_id", orderID.Hex()))
		return nil, err
	}

	return order.Items, nil
}

// FindOrdersWithFilter retrieves orders with custom filters and pagination
func (r *OrderRepository) FindOrdersWithFilter(ctx context.Context, filter bson.M, page, limit int) ([]models.Order, int64, error) {
	logger.Logger.Infof("Repository - FindOrdersWithFilter %v", filter, logger.Int("page", page), logger.Int("limit", limit))

	// Đếm tổng số đơn hàng phù hợp với bộ lọc
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Logger.Errorf("Failed to count filtered orders: %v", err)
		return nil, 0, err
	}
	logger.Logger.Info("Repository - Total documents matching filter")

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
		logger.Err("Failed to find filtered orders", err, logger.Int("page", page), logger.Int("limit", limit))
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		logger.Err("Failed to decode filtered orders", err, logger.Int("page", page), logger.Int("limit", limit))
		return nil, 0, err
	}
	logger.Info("Repository - Found filtered orders", logger.Int("count", len(orders)))

	return orders, total, nil
}

// GetOrderItemDetails retrieves detailed information for items in an order
// This replicates the aggregation pipeline from the commented code
func (r *OrderRepository) GetOrderItemDetails(ctx context.Context, orderID primitive.ObjectID) ([]models.OrderItem, error) {
	logger.Info("Repository - GetOrderItemDetails", logger.Str("order_id", orderID.Hex()))

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
		logger.Err("Error aggregating order items", err, logger.Str("order_id", orderID.Hex()))
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []models.OrderItem
	if err := cursor.All(ctx, &items); err != nil {
		logger.Err("Error decoding order items", err, logger.Str("order_id", orderID.Hex()))
		return nil, err
	}
	logger.Info("Repository - Found order item details", logger.Int("count", len(items)), logger.Str("order_id", orderID.Hex()))

	return items, nil
}

// CalculateOrderPages calculates pagination info based on total orders and limit
func CalculateOrderPages(total int64, limit int) int {
	return int(math.Ceil(float64(total) / float64(limit)))
}

// UpdateOrderStatus updates the status of an order
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID primitive.ObjectID, status string) error {
	logger.Info("Repository - UpdateOrderStatus", logger.Str("order_id", orderID.Hex()), logger.Str("status", status))

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": orderID}, update)
	if err != nil {
		logger.Err("Failed to update order status", err, logger.Str("order_id", orderID.Hex()), logger.Str("status", status))
		return err
	}

	return nil
}

// FindOrderByID retrieves a specific order by ID
func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID primitive.ObjectID) (*models.Order, error) {
	logger.Info("Repository - GetOrderByID", logger.Str("order_id", orderID.Hex()))

	var order models.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Err("Order not found", err, logger.Str("order_id", orderID.Hex()))
			return nil, nil
		}
		logger.Err("Error finding order", err, logger.Str("order_id", orderID.Hex()))
		return nil, err
	}

	return &order, nil
}
