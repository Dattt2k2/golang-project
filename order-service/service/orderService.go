package service

import (
	"context"
	"fmt"
	"time"

	// "net/http"

	"github.com/Dattt2k2/golang-project/order-service/models"
	"github.com/Dattt2k2/golang-project/product-service/gRPC/service"
	// "github.con/Dattt2k2/golang-project/order-service/database"
	"go.mongodb.org/mongo-driver/mongo"
	// "github.com/gin-gonic/gin"
)


var orderCollection *mongo.Collection

func CheckStock(ctx context.Context, orderItems []models.OrderItem) error {

	productClient  := ProductServiceConnection() 

	var stockRequest service.StockRequest
	for _, item := range orderItems {
		stockRequest.Items = append(stockRequest.Items, &service.StockItem{
			ProductId: item.ProductID.String(),
			Quantity: int32(item.Quantity),
		})
	}

	stockResponse, err := productClient.CheckStock(ctx, &stockRequest)
	if err != nil{
		return fmt.Errorf("Error checking stock: %v", err)
	}

	for _, status := range  stockResponse.Status{
		if !status.InStock {
			return fmt.Errorf("Product %s is out of stock", status.ProductId)
		}
	}
	return nil
}

func UpdateStock(ctx context.Context, orderItems []models.OrderItem) error{

	productClient := ProductServiceConnection()

	var updateRequest service.UpdateStockRequest
	for _, item := range orderItems {
		updateRequest.Items = append(updateRequest.Items, &service.StockItem{
			ProductId: item.ProductID.String(),
			Quantity:  int32(item.Quantity),
		})
	}

	updateRsonse, err := productClient.UpdateStock(ctx, &updateRequest)
	if err != nil{
		return fmt.Errorf("Error updating stock: %v", err)
	}


	for _, status := range updateRsonse.UpdateStatus{
		if !status.Updated{
			return fmt.Errorf("Failed to update stock for product %s: %s", status.ProductId, status.Message)
		}
	}
	return nil
}


func SaveOrderToDB(ctx context.Context, orderItems []models.OrderItem) error{
	order := models.Order{Items: orderItems, Created_at: time.Now()}

	_, err := orderCollection.InsertOne(ctx, order)
	if err != nil {
		return fmt.Errorf("Failed to save order %v", err)
	}

	select{
	case <-ctx.Done():
		return fmt.Errorf("Operation timed out: %v", ctx.Err())
	default:

	}
	return nil

}