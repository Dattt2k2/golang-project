package controller

import (
	"context"

	pb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
)

type CartServer struct{
	pb.UnimplementedCartServiceServer
}

func (s *CartServer) GetCartItems(ctx context.Context, req * pb.CartRequest) (*pb.CartResponse, error){
	id := re
}