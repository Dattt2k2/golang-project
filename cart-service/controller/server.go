// package controller

// import (
// 	// "context"
// 	// "log"

// 	// "github.com/Dattt2k2/golang-project/cart-service/models"
// 	"context"
// 	"log"

// 	"github.com/Dattt2k2/golang-project/cart-service/models"
// 	pb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// 	// "go.mongodb.org/mongo-driver/bson/primitive"
// 	// "google.golang.org/grpc/codes"
// 	// "google.golang.org/grpc/status"
// )

// type CartServer struct {
// 	pb.UnimplementedCartServiceServer
// }

// // func (s *CartServer) GetCartItems(ctx context.Context, req * pb.CartRequest) (*pb.CartResponse, error){
// // 	id := req.UserId

// // 	productID, err := primitive.ObjectIDFromHex(id)
// // 	if err != nil{
// // 		return nil, status.Errorf(codes.InvalidArgument, "Invalid product ID formate: %v", err)

// // 	}

// // 	log.Printf("product id: %v", productID)

// // 	var cart models.CartItem

// // 	if err := cartCollection.FindOne()
// // }

// func (s *CartServer) GetCartItems(ctx context.Context, req *pb.CartRequest) (*pb.CartResponse, error) {
// 	userId := req.UserId
// 	log.Printf("User ID: %v", userId)

// 	if userId == "" {
// 		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
// 	}

// 	userObjectId, err := primitive.ObjectIDFromHex(userId)
// 	if err != nil {
// 		return nil, status.Errorf(codes.InvalidArgument, "Invalid User ID format: %v", err)
// 	}

// 	var cart models.Cart

// 	err = cartCollection.FindOne(ctx, bson.M{"user_id": userObjectId}).Decode(&cart)
// 	if err != nil {
// 		log.Printf("Error finding cart: %v", err)
// 		return nil, status.Errorf(codes.NotFound, "Cart not found for user ID: %v", userId)
// 	}

// 	var items []*pb.CartItem
// 	for _, item := range cart.Items {
// 		cartItem := &pb.CartItem{
// 			ProductId: item.ProductID.Hex(),
// 			Quantity:  int32(item.Quantity),
// 			Price:     float32(item.Price),
// 			Name:      item.Name,
// 		}
// 		items = append(items, cartItem)
// 	}

// 	response := &pb.CartResponse{
// 		Items: items,
// 	}

// 	log.Printf("Cart response: %v", response)
// 	return response, nil

// }

package controller

import (
	"context"
	"log"

	"github.com/Dattt2k2/golang-project/cart-service/service"
	pb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
) 


type CartServer struct {
	pb.UnimplementedCartServiceServer
	cartService service.CartService
}

func NewCartServer(cartService service.CartService) *CartServer {
	return &CartServer{
		cartService: cartService,

	}
}

func (s *CartServer) GetCartItems (ctx context.Context, req *pb.CartRequest) (*pb.CartResponse, error) {
	userID := req.UserId 
	
	if userID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
	}

	cart, err := s.cartService.GetUserCart(ctx, userID)
	if err != nil {
		log.Printf("Error getting cart: %v", err)
		return nil, status.Errorf(codes.Internal, "Cart not found: %v", err)
	}

	var items []*pb.CartItem
	for _, item := range cart.Items {
		cartItem := &pb.CartItem{
			ProductId: item.ProductID.Hex(),
			Quantity:  int32(item.Quantity),
			Price:     float32(item.Price),
			Name:      item.Name,
		}
		items = append(items, cartItem)
	}

	response := &pb.CartResponse{
		Items: items, 
	}

	log.Printf("Cart response: %v", response)
	return response, nil
}