package service

import (
	"log"

	// pb "github.com/Dattt2k2/golang-project/order-service/gRPC/service"
    cartPb "github.com/Dattt2k2/golang-project/cart-service/gRPC/service"
    productPb "github.com/Dattt2k2/golang-project/product-service/gRPC/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CartServiceConnection() cartPb.CartServiceClient {
	conn, err := grpc.NewClient("cart-service:8082", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil{
		log.Fatalf("Failed to connect to Cart-serviceL %v", err)
	}

	log.Println("Connected to Cart-service")
	return cartPb.NewCartServiceClient(conn)
}

func ProductServiceConnection() productPb.ProductServiceClient {
	conn, err := grpc.NewClient("product-service:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil{
		log.Fatalf("Failed to connect to Product-service: %v", err )
	}

	log.Println("Connected to Product-service")
	return productPb.NewProductServiceClient(conn)
}