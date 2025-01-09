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

func (s *ProductServer) GetBasicInfo(ctx context.Context, req *pb.ProductRequest) (*pb.BasicProductResponse, error){

	id := req.Id

	var product models.Product


	if err := productCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&product); err != nil{
		return nil, err
	}

	return &pb.BasicProductResponse{
		Id: product.ID.String(),
		Name: *product.Name,
		Price: float32(product.Price),
	}, nil
}

func (s *ProductServer) GetProductInfo(ctx context.Context, req *pb.ProductRequest) (*pb.ProductResponse, error){
	
	id := req.Id

	var product models.Product

	if err := productCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&product); err != nil{
		return nil, err
	}
	
	return &pb.ProductResponse{
		Id: product.ID.String(),
		Name : *product.Name,
		Description: *product.Description,
		Price: float32(product.Price),
		Quantity: int32(*product.Quantity),
		ImageUrl: product.ImagePath,
	}, nil
}

func (s *ProductServer) CheckStock(ctx context.Context, req *pb.ProductRequest) (*pb.StockResponse, error){
	
	id := req.Id

	var product models.Product
	if err := productCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&product); err != nil{
		return nil, err
	}

	if *product.Quantity > 0 {
		
		return &pb.StockResponse{
			InStock: true,
			AvailableQuantity: int32(*product.Quantity),
			Message: "Product is in stock",
		}, nil
	}

	return &pb.StockResponse{
		InStock: false,
		AvailableQuantity: int32(*product.Quantity),
		Message: "Product is out of stock",
	}, nil
}
