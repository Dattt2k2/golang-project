package service

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"

	orderPb "github.com/Dattt2k2/golang-project/module/gRPC-Order/service"
)

// NewOrderServiceClient creates a new gRPC client for the OrderService.
func NewOrderServiceClient(address string) (orderPb.OrderServiceClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to Order service: %v", err)
		return nil, err
	}
	return orderPb.NewOrderServiceClient(conn), nil
}