package service

import (
	"context"
	pb "module/gRPC-Order/service"
	"order-service/repositories"
)

type OrderServiceServer struct {
	pb.UnimplementedOrderServcieServer
	orderRepo repositories.OrderRepository
}

func (s *OrderServiceServer) HasPurchased(ctx context.Context, req *pb.HasPurchasedRequest) (*pb.HasPurchasedResponse, error) {
	userID := req.GetUserId()
	productID := req.GetProductId()

	_, err := s.orderRepo.GetUserOrderWithProductID(ctx, userID, productID)
	if err != nil {
		if err.Error() == "record not found" {
			return &pb.HasPurchasedResponse{Purchased: false}, nil
		}
		return nil, err
	}		

	return &pb.HasPurchasedResponse{Purchased: true}, nil  
}