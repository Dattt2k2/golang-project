package service

import (

	// pb "github.com/Dattt2k2/golang-project/order-service/gRPC/service"
    cartPb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
    productPb "github.com/Dattt2k2/golang-project/module/gRPC-Product/service"
	"github.com/Dattt2k2/golang-project/order-service/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CartServiceConnection() cartPb.CartServiceClient {
	conn, err := grpc.NewClient("cart-service:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil{
		logger.Err("Failed to connect to Cart-service", err)
	}
	
	logger.Info("Connected to Cart-service")
	return cartPb.NewCartServiceClient(conn)
}

func ProductServiceConnection() productPb.ProductServiceClient {
	conn, err := grpc.NewClient("product-service:8089", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil{
		logger.Err("Failed to connect to Product-service", err)
	}

	if err != nil{
		logger.Err("Failed to connect to Product-service", err)
	}

	logger.Info("Connected to Product-service")
	return productPb.NewProductServiceClient(conn)
}