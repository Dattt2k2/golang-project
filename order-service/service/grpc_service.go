package service

import (
	"context"
	pb "module/gRPC-Order/service"
	logger "order-service/log"
	"order-service/repositories"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderServiceServer struct {
	pb.UnimplementedOrderServiceServer
	OrderRepo *repositories.OrderRepository
}

func (s *OrderServiceServer) HasPurchased(ctx context.Context, req *pb.HasPurchasedRequest) (*pb.HasPurchasedResponse, error) {
	// Add timeout to prevent hanging requests
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Validate request
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	userID := req.GetUserId()
	productID := req.GetProductId()

	if userID == "" || productID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and product_id are required")
	}

	// Use a channel to handle timeout properly
	type result struct {
		purchased bool
		err       error
	}

	resultCh := make(chan result, 1)

	go func() {
		_, err := s.OrderRepo.GetUserOrderWithProductID(ctx, userID, productID)
		if err != nil {
			if err.Error() == "record not found" {
				resultCh <- result{purchased: false, err: nil}
				return
			}
			resultCh <- result{purchased: false, err: err}
			return
		}
		resultCh <- result{purchased: true, err: nil}
	}()

	select {
	case <-ctx.Done():
		logger.Logger.Warn("HasPurchased request timeout or cancelled",
			"user_id", userID,
			"product_id", productID,
		)
		return nil, status.Error(codes.DeadlineExceeded, "request timeout")
	case res := <-resultCh:
		if res.err != nil {
			logger.Err("Failed to check purchase history", res.err)
			return nil, status.Error(codes.Internal, "failed to check purchase history")
		}
		return &pb.HasPurchasedResponse{Purchased: res.purchased}, nil
	}
}
