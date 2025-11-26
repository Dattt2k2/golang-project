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
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CartService interface {
    AddToCart(ctx context.Context, userID string, productID string, quantity int) error
    GetUserCart(ctx context.Context, userID string) (*models.Cart, error)
    DeleteProductFromCart(ctx context.Context, userID string, productID string) error
    ClearCart(ctx context.Context, userID string) error
    GetAllCarts(ctx context.Context, page, limit int) ([]models.Cart, int, int, bool, bool, error)
	UpdateCartItem(ctx context.Context, userID string, productID string, quantity int) error
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

func (s *cartServiceImpl) AddToCart(ctx context.Context, userID string, productID string, quantity int ) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := uuid.Parse(userID); err != nil {
		return errors.New("invalid User ID format")
	}
	if _, err := uuid.Parse(productID); err != nil {
		return errors.New("invalid Product ID format")
	}


	productReq := &pb.ProductRequest{
		Id: productID,
	}
	basicInfo, err := s.productClient.GetProductInfo(ctx, productReq)
	if err != nil {
		log.Printf("Failed to get product info: %v", err)
		return errors.New("failed to get product info")
	}

	if basicInfo.VendorId == userID {
		return errors.New("cannot add your own product to cart")
	}

	checkStock, err := s.productClient.CheckStock(ctx, productReq)
	if err != nil {
		return errors.New("failed to check product stock")
	}
	avaiableQuantity := int(checkStock.AvailableQuantity)
	if quantity > avaiableQuantity {
		return errors.New("not enough stock available")
	}


	cartItem := models.CartItem{
		VendorID: basicInfo.VendorId,
		ProductID: productID,
		Name: basicInfo.Name,
		Price: float64(basicInfo.Price),
		Quantity: quantity,
		ImageUrl: basicInfo.ImageUrl,
	}

	return s.repo.AddItem(ctx, userID, cartItem)
}


func (s *cartServiceImpl) GetUserCart(ctx context.Context, userID string) (*models.Cart, error){
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := uuid.Parse(userID); err != nil {
		return nil, errors.New("Invalid User ID format")
	}

	return s.repo.FindByUserID(ctx, userID)

}


func (s *cartServiceImpl) DeleteProductFromCart(ctx context.Context, userID string, productID string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := uuid.Parse(userID); err != nil {
		return errors.New("Invalid User ID format")
	}

	if _, err := uuid.Parse(productID); err != nil {
		return errors.New("Invalid Product ID format")
	}
	modifiedCount, err := s.repo.RemoveItem(ctx, userID, productID)
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

	if _, err := uuid.Parse(userID); err != nil {
		return errors.New("Invalid User ID format")
	}

	return s.repo.ClearCart(ctx, userID)
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

func (s *cartServiceImpl) UpdateCartItem(ctx context.Context, userID string, productID string, quantity int) error {
	if _, err := uuid.Parse(userID); err != nil {
		return errors.New("invalid User ID format")
	}

	if _, err := uuid.Parse(productID); err != nil {
		return errors.New("invalid Product ID format")
	}

	return s.repo.UpdateCartItem(ctx, userID, productID, quantity)
}