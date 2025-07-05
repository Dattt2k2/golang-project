package service

import (
	"context"
	"errors"
	"log"
	"math"
	"time"

	"cart-service/models"
	"cart-service/repository"
	pb "github.com/Dattt2k2/golang-project/module/gRPC-Product/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CartService interface {
    AddToCart(ctx context.Context, userID string, productID string, quantity int) error
    GetUserCart(ctx context.Context, userID string) (*models.Cart, error)
    DeleteProductFromCart(ctx context.Context, userID string, productID string) error
    ClearCart(ctx context.Context, userID string) error
    GetAllCarts(ctx context.Context, page, limit int) ([]models.Cart, int, int, bool, bool, error)
}

type cartServiceImpl struct {
	repo repository.CartRepository
	productClient pb.ProductServiceClient
}

func NewCartService(repo repository.CartRepository) (CartService, error) {
	conn, err := grpc.NewClient("product-service:8089", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to product service: %v", err)
		return nil, err 
	}

	log.Printf("Connected to product service")
	productClient := pb.NewProductServiceClient(conn)

	return &cartServiceImpl{
		repo: repo,
		productClient: productClient,
	}, nil 
}

func (s *cartServiceImpl) AddToCart(ctx context.Context, UserID string, productID string, quantity int ) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	userObj, err := primitive.ObjectIDFromHex(UserID)
	if err != nil {
		return errors.New("Invalid user ID format")
	}

	productReq := &pb.ProductRequest{
		Id: productID,
	}
	basicInfo, err := s.productClient.GetBasicInfo(ctx, productReq)
	if err != nil {
		log.Printf("Failed to get product info: %v", err)
		return errors.New("Failed to get product info")
	}


	checkStock, err := s.productClient.CheckStock(ctx, productReq)
	if err != nil {
		return errors.New("Failed to check product stock")
	}
	avaiableQuantity := int(checkStock.AvailableQuantity)
	if quantity > avaiableQuantity {
		return errors.New("Not enough stock available")
	}

	productObj, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return errors.New("Invalid product ID format")
	}


	cartItem := models.CartItem{
		ProductID: productObj,
		Name: basicInfo.Name,
		Price: float64(basicInfo.Price),
		Quantity: quantity,
	}

	return s.repo.AddItem(ctx, userObj, cartItem)
}


func (s *cartServiceImpl) GetUserCart(ctx context.Context, userID string) (*models.Cart, error){
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("Invalid user ID format")
	}

	return s.repo.FindByUserID(ctx, userObjID)

}


func (s *cartServiceImpl) DeleteProductFromCart(ctx context.Context, userID string, productID string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("Invalid user ID")
	}

	productObjID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return errors.New("Invalid product ID")
	}

	modifiedCount, err := s.repo.RemoveItem(ctx,  userObjID, productObjID)
	if err != nil {
		return errors.New("Failed to remove item from cart")
	}

	if modifiedCount == 0 {
		return errors.New("No item found in cart")
	}

	return nil
}


func (s *cartServiceImpl) ClearCart(ctx context.Context, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("Invalid user ID format")
	}

	return s.repo.ClearCart(ctx, userObjID)
}

func (s *cartServiceImpl) GetAllCarts(ctx context.Context, page, limit int) ([]models.Cart, int, int, bool, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	carts, total, err := s.repo.GetAllCarts(ctx, page, limit)
	if err != nil {
		return nil, 0,0, false, false, err
	}

	pages := int(math.Ceil(float64(total) / float64(limit)))
	hasNext := page < pages 
	hasPrevious := page > 1

	for i, cart := range carts {
		items, err := s.repo.GetCartItems(ctx, cart.ID)
		if err != nil {
			log.Printf("Failed to get cart items: %v", err)
			continue 
		}
		carts[i].Items = items 
	}

	return carts, int(total), pages, hasNext, hasPrevious, nil
}