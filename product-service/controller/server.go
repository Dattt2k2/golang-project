package controllers

import (
	"context"
	"log"
	"strings"

	pb "module/gRPC-Product/service"
	logger "product-service/log"
	"product-service/service"

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

	product, err := s.service.GetProductByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Product not found: %v", err)
	}

	// Lấy giá từ variant đầu tiên (nếu có)
	var price float32
	if len(product.Variants) > 0 {
		price = float32(product.Variants[0].Price)
	}

	return &pb.BasicProductResponse{
		Id:    product.ID,
		Name:  product.Name,
		Price: price,
	}, nil
}

func (s *ProductServer) GetProductInfo(ctx context.Context, req *pb.ProductRequest) (*pb.ProductResponse, error) {
	id := req.Id

	product, err := s.service.GetProductByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Product not found: %v", err)
	}

	imageUrls := ""
	if len(product.ImagePath) > 0 {
		imageUrls = strings.Join(product.ImagePath, ",")
	}

	// Chuyển đổi tất cả variants sang protobuf format
	pbVariants := make([]*pb.ProductVariant, len(product.Variants))
	for i, v := range product.Variants {
		pbVariants[i] = &pb.ProductVariant{
			Id:        v.ID,
			Size:      v.Size,
			Color:     v.Color,
			Material:  v.Material,
			CostPrice: float32(v.CostPrice),
			Price:     float32(v.Price),
			Quantity:  int32(v.Quantity),
		}
	}

	return &pb.ProductResponse{
		Id:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		ImageUrl:    imageUrls,
		VendorId:    product.UserID,
		Category:    product.Category,
		Status:      product.Status,
		Variants:    pbVariants,
	}, nil
}

func (s *ProductServer) GetBasicInfo(ctx context.Context, req *pb.ProductRequest) (*pb.BasicProductResponse, error) {
	id := req.Id
	product, err := s.service.GetProductByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Product not found: %v", err)
	}
	log.Printf("product id: %v", id)

	// Lấy giá từ variant đầu tiên (nếu có)
	var price float32
	if len(product.Variants) > 0 {
		price = float32(product.Variants[0].Price)
	}

	return &pb.BasicProductResponse{
		Id:       product.ID,
		Name:     product.Name,
		Price:    price,
		VendorId: product.UserID,
	}, nil
}

func (s *ProductServer) CheckStock(ctx context.Context, req *pb.ProductRequest) (*pb.StockResponse, error) {
	id := req.Id
	product, err := s.service.GetProductByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Product not found: %v", err)
	}

	// Tính tổng số lượng từ tất cả variants
	var totalQuantity int32
	for _, v := range product.Variants {
		totalQuantity += int32(v.Quantity)
	}

	if totalQuantity > 0 {
		return &pb.StockResponse{
			InStock:           true,
			AvailableQuantity: totalQuantity,
			Message:           "Product is in stock",
		}, nil
	}

	return &pb.StockResponse{
		InStock:           false,
		AvailableQuantity: totalQuantity,
		Message:           "Product is out of stock",
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
		imageUrls := ""
		if len(p.ImagePath) > 0 {
			imageUrls = strings.Join(p.ImagePath, ",")
		}

		// Lấy giá từ variant đầu tiên (nếu có)
		var price float32
		if len(p.Variants) > 0 {
			price = float32(p.Variants[0].Price)
		}

		pbProducts = append(pbProducts, &pb.Product{
			Id:          p.ID,
			Name:        p.Name,
			Price:       price,
			Description: p.Description,
			ImageUrl:    imageUrls,
			Category:    p.Category,
		})
	}
	return &pb.ProductList{Products: pbProducts}, nil
}

func (s *ProductServer) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	var updateStatus []*pb.StockUpdateStatus
	allSuccess := true

	for _, item := range req.Items {
		err := s.service.UpdateProductStock(ctx, item.ProductId, item.VariantId, int(item.Quantity))

		status := &pb.StockUpdateStatus{
			ProductId: item.ProductId,
			Updated:   err == nil,
		}

		if err != nil {
			allSuccess = false
			status.Message = err.Error()
		} else {
			status.Message = "Stock updated successfully"
		}

		updateStatus = append(updateStatus, status)
	}

	return &pb.UpdateStockResponse{
		UpdateStatus: updateStatus,
		Success:      allSuccess,
		Message:      "Stock update completed",
	}, nil
}
