package controllers

import (
	"context"
	"log"

	"product-service/log"
	pb "github.com/Dattt2k2/golang-project/module/gRPC-Product/service"
	"product-service/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductServer struct {
	pb.UnimplementedProductServiceServer
	service service.ProductService
}

// func (s *ProductServer) GetBasicInfo(ctx context.Context, req *pb.ProductRequest) (*pb.BasicProductResponse, error){

// 	id := req.Id
// 	log.Printf("Product id: %v", id)
// 	productID, err := primitive.ObjectIDFromHex(id)
// 	if err != nil {
// 		return nil, status.Errorf(codes.InvalidArgument, "Invalid product ID format: %v", err)
// 	}



// 	log.Printf("product id: %v", productID)
// 	var product models.Product


// 	if err := productCollection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product); err != nil{
// 		return nil, err
// 	}

// 	return &pb.BasicProductResponse{
// 		Id: product.ID.String(),
// 		Name: *product.Name,
// 		Price: float32(product.Price),
// 	}, nil
// }

// func (s *ProductServer) GetProductInfo(ctx context.Context, req *pb.ProductRequest) (*pb.ProductResponse, error){
// 	id := req.Id
// 	log.Printf("Product id: %v", id)
// 	productID, err := primitive.ObjectIDFromHex(id)
// 	if err != nil {
// 		return nil, status.Errorf(codes.InvalidArgument, "Invalid product ID format: %v", err)
// 	}



// 	log.Printf("product id: %v", productID)
// 	var product models.Product


// 	if err := productCollection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product); err != nil{
// 		return nil, err
// 	}

// 	return &pb.ProductResponse{
// 		Id: product.ID.String(),
// 		Name: *product.Name,
// 		Price: float32(product.Price),
// 		Description: *product.Description,
// 		ImageUrl: product.ImagePath,
// 		Quantity: int32(*product.Quantity),
// 	}, nil
// }

// func (s *ProductServer) CheckStock(ctx context.Context, req *pb.ProductRequest) (*pb.StockResponse, error){
	
// 	id := req.Id

// 	log.Printf("Product id: %v", id)
// 	productID, err := primitive.ObjectIDFromHex(id)
// 	if err != nil {
// 		return nil, status.Errorf(codes.InvalidArgument, "Invalid product ID format: %v", err)
// 	}

// 	var product models.Product
// 	if err := productCollection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product); err != nil{
// 		return nil, err
// 	}

// 	if *product.Quantity > 0 {
		
// 		return &pb.StockResponse{
// 			InStock: true,
// 			AvailableQuantity: int32(*product.Quantity),
// 			Message: "Product is in stock",
// 		}, nil
// 	}

// 	return &pb.StockResponse{
// 		InStock: false,
// 		AvailableQuantity: int32(*product.Quantity),
// 		Message: "Product is out of stock",
// 	}, nil
// }



func NewProductServer(service service.ProductService) *ProductServer {
	return &ProductServer{
		service: service,
	}
}

func (s *ProductServer) AddProduct(ctx context.Context, req *pb.ProductRequest) (*pb.BasicProductResponse, error) {
	id := req.Id 
	log.Printf("Product id: %v", id)
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid product ID format: %v", err)
	}

	log.Printf("product id: %v", productID)

	product, err := s.service.GetProductByID(ctx, productID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Product not found: %v", err)
	}

	return &pb.BasicProductResponse{
		Id: product.ID.String(),
		Name: product.Name,
		Price: float32(product.Price),
	}, nil
}

func (s *ProductServer) GetProductInfo(ctx context.Context, req *pb.ProductRequest) (*pb.ProductResponse, error) {
	id := req.Id
	log.Printf("Product id: %v", id)
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid product ID format: %v", err)
	}
	log.Printf("product id: %v", productID)

	product, err := s.service.GetProductByID(ctx, productID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Product not found: %v", err)
	}

	return &pb.ProductResponse{
		Id: product.ID.String(),
		Name: product.Name,
		Price: float32(product.Price),
		Description: product.Description,
		ImageUrl: product.ImagePath,
		Quantity: int32(product.Quantity),
	}, nil 
}

func (s *ProductServer) GetBasicInfo(ctx context.Context, req *pb.ProductRequest) (*pb.BasicProductResponse, error){
	id := req.Id 
	log.Printf("Product id: %v", id)
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid product ID format: %v", err)
	}

	product, err := s.service.GetProductByID(ctx, productID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Product not found: %v", err)
	}
	log.Printf("product id: %v", productID)
	return &pb.BasicProductResponse{
		Id: product.ID.Hex(),
		Name: product.Name,
		Price: float32(product.Price),

	}, nil
}

func (s *ProductServer) CheckStock(ctx context.Context, req *pb.ProductRequest) (*pb.StockResponse, error) {
	id := req.Id 
	log.Printf("Product id: %v", id)
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid product ID format: %v", err)
	}

	product, err := s.service.GetProductByID(ctx, productID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Product not found: %v", err)
	}

	if product.Quantity > 0 {
		return &pb.StockResponse{
			InStock: true,
			AvailableQuantity: int32(product.Quantity),
			Message: "Product is in stock",
		}, nil
	}

	return &pb.StockResponse{
		InStock: false,
		AvailableQuantity: int32(product.Quantity),
		Message: "Product is out of stock",
	}, nil
}

// GetAllProduct for  re-indexes products in Elasticsearch 
func (s *ProductServer) GetAllProducts(ctx context.Context, req *pb.Empty) (*pb.ProductList, error) {
	products, err := s.service.GetAllProductForIndex(ctx)
	if err != nil {
		logger.Err("Failed to get products", err)
		return nil, status.Errorf(codes.Internal, "Failed to get products: %v", err)
	}

	var pbProducts []*pb.Product 
	for _, p := range products {
		pbProducts = append(pbProducts, &pb.Product{
			Id: p.ID.Hex(),
			Name: p.Name,
			Price: float32(p.Price),
			Description: p.Description,
			ImageUrl: p.ImagePath,
			Category: p.Category,
		})
	}
	return &pb.ProductList{Products: pbProducts}, nil
}