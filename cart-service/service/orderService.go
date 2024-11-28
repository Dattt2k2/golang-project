package service

import "github.com/Dattt2k2/golang-project/cart-service/repositories"

type OrderService struct{
	orderRepo		*repositories.OrderRepo
}

func NewOrderService(orderRepo *repositories.OrderRepo) *OrderService {
	return &OrderService{orderRepo: orderRepo}
}

func 