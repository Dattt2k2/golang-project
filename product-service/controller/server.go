package controllers

import (
	"context"

	pb "github.com/Dattt2k2/golang-project/module/gRPC-Product/service"
	"github.com/Dattt2k2/golang-project/product-service/models"
	"go.mongodb.org/mongo-driver/bson"
)

type ProductServer struct {
	pb.UnimplementedProductServiceServer
}

func (s *ProductServer) GetProductInfor(ctx context.Context, req *pb.ProductRequest) (*pb.ProductResponse, error){
	
	id := req.Id

	var product models.Product

	if err := productCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&product); err != nil{
		return nil, err
	}
	
	return &pb.ProductResponse{
		Id: product.ID.Hex(),
		Name : *product.Name,
		Description: *product.Description,
		Price: float32(product.Price),
		Quantity: int32(*product.Quantity),
	}, nil
}
