package controller

import (
	// "context"
	// "log"

	// "github.com/Dattt2k2/golang-project/cart-service/models"
	pb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/status"
)

type CartServer struct{
	pb.UnimplementedCartServiceServer
}

// func (s *CartServer) GetCartItems(ctx context.Context, req * pb.CartRequest) (*pb.CartResponse, error){
// 	id := req.UserId

// 	productID, err := primitive.ObjectIDFromHex(id)
// 	if err != nil{
// 		return nil, status.Errorf(codes.InvalidArgument, "Invalid product ID formate: %v", err)

// 	}

// 	log.Printf("product id: %v", productID)

// 	var cart models.CartItem

// 	if err := cartCollection.FindOne()
// }