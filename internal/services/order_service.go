package services

import (
	"context"

	"GameStoreAPI/internal/repository"
)

type OrderService struct {
	Repo *repository.OrderRepository
}

func NewOrderService(r *repository.OrderRepository) *OrderService {
	return &OrderService{Repo: r}
}

func (s *OrderService) Checkout(ctx context.Context, customerID int64) (int64, error) {
	return s.Repo.CreateOrderFromCart(ctx, customerID)
}
