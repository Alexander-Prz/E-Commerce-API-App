package services

import (
	"GameStoreAPI/internal/repository"
	"context"
)

type CustomerGamesService struct {
	Repo     *repository.CustomerGamesRepository
	CartRepo *repository.CartRepository // for order items
}

func NewCustomerGamesService(r *repository.CustomerGamesRepository, cart *repository.CartRepository) *CustomerGamesService {
	return &CustomerGamesService{Repo: r, CartRepo: cart}
}

// Called by checkout flow
func (s *CustomerGamesService) InsertOwnership(ctx context.Context, customerID int64, gameIDs []int64) error {
	return s.Repo.InsertPurchased(ctx, customerID, gameIDs)
}

func (s *CustomerGamesService) ListOwned(ctx context.Context, customerID int64) (interface{}, error) {
	return s.Repo.ListOwnedGames(ctx, customerID)
}

func (s *CustomerGamesService) ListOrders(ctx context.Context, customerID int64) (interface{}, error) {
	return s.Repo.ListOrders(ctx, customerID)
}

func (s *CustomerGamesService) OrderDetails(ctx context.Context, customerID, orderID int64) (interface{}, interface{}, error) {
	return s.Repo.GetOrderDetails(ctx, customerID, orderID)
}

func (s *CustomerGamesService) ListAllOrders(ctx context.Context) (interface{}, error) {
	return s.Repo.ListAllOrders(ctx)
}

func (s *CustomerGamesService) GetOrderDetailsAdmin(ctx context.Context, orderID int64) (interface{}, interface{}, error) {
	return s.Repo.GetOrderDetailsAdmin(ctx, orderID)
}
