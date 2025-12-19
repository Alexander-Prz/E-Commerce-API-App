package services

import (
	"context"
	"errors"
	"fmt"

	"GameStoreAPI/internal/model"
	"GameStoreAPI/internal/repository"
)

type CartService struct {
	Repo              *repository.CartRepository
	OrderRepo         *repository.OrderRepository
	CustomerGamesRepo *repository.CustomerGamesRepository
	AuthRepo          *repository.AuthRepository
	CustomerRepo      *repository.CustomerRepository
}

func NewCartService(r *repository.CartRepository, or *repository.OrderRepository, cgr *repository.CustomerGamesRepository, ar *repository.AuthRepository, cr *repository.CustomerRepository) *CartService {
	return &CartService{
		Repo:              r,
		OrderRepo:         or,
		CustomerGamesRepo: cgr,
		AuthRepo:          ar,
		CustomerRepo:      cr,
	}
}

func (s *CartService) Add(ctx context.Context, authID, gameID int64) error {
	// resolve customer
	cid, err := s.Repo.GetCustomerID(ctx, authID)
	if err != nil {
		return err
	}

	// 🚨 block owned games
	owned, err := s.CustomerGamesRepo.ExistsAnyOwned(ctx, cid, []int64{gameID})
	if err != nil {
		return err
	}
	if owned != 0 {
		return errors.New("game already owned")
	}

	// get or create open cart
	orderID, err := s.Repo.FindOpenOrder(ctx, cid)
	if err != nil {
		orderID, err = s.Repo.CreateOpenOrder(ctx, cid)
		if err != nil {
			return err
		}
	}

	// 🚨 block duplicate cart entry
	exists, err := s.Repo.IsGameInCart(ctx, orderID, gameID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("game already in your cart")
	}

	// get price
	_, price, err := s.Repo.GetGameInfo(ctx, gameID)
	if err != nil {
		return err
	}

	// insert single item (qty = 1)
	return s.Repo.InsertOrderItem(ctx, orderID, gameID, price)
}

/* ARTIFACT
// Update sets quantity for an item in the cart
func (s *CartService) Update(ctx context.Context, authID, gameID int64, qty int) error {
	if qty <= 0 {
		return errors.New("quantity must be > 0")
	}
	cid, err := s.Repo.GetCustomerID(ctx, authID)
	if err != nil {
		return err
	}
	orderID, err := s.Repo.FindOpenOrder(ctx, cid)
	if err != nil {
		return errors.New("no open cart")
	}
	return s.Repo.SetOrderItemQuantity(ctx, orderID, gameID, qty)
}
*/

// Remove removes an item from the cart
func (s *CartService) Remove(ctx context.Context, authID, gameID int64) error {
	cid, err := s.Repo.GetCustomerID(ctx, authID)
	if err != nil {
		return err
	}
	orderID, err := s.Repo.FindOpenOrder(ctx, cid)
	if err != nil {
		return errors.New("no open cart")
	}
	return s.Repo.RemoveOrderItem(ctx, orderID, gameID)
}

// Clear clears the cart (removes items)
func (s *CartService) Clear(ctx context.Context, authID int64) error {
	cid, err := s.Repo.GetCustomerID(ctx, authID)
	if err != nil {
		return err
	}
	orderID, err := s.Repo.FindOpenOrder(ctx, cid)
	if err != nil {
		return errors.New("no open cart")
	}
	return s.Repo.ClearOrderItems(ctx, orderID)
}

// Get returns the cart (items + total)
func (s *CartService) Get(ctx context.Context, authID int64) (*model.CartResponse, error) {
	cid, err := s.Repo.GetCustomerID(ctx, authID)
	if err != nil {
		return nil, err
	}
	orderID, err := s.Repo.FindOpenOrder(ctx, cid)
	if err != nil {
		// empty cart
		return &model.CartResponse{Items: []model.CartItem{}, Total: 0}, nil
	}
	items, total, err := s.Repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, err
	}
	resp := &model.CartResponse{
		Items: items,
		Total: total,
	}
	return resp, nil
}

func (s *CartService) Checkout(ctx context.Context, authID int64) (int64, error) {
	// check user exists and not banned
	u, err := s.AuthRepo.GetByID(ctx, authID)
	if err != nil {
		return 0, err
	}
	if u.DeletedAt != nil {
		return 0, errors.New("user is banned")
	}

	// get customer id
	cid, err := s.Repo.GetCustomerID(ctx, authID)
	if err != nil {
		return 0, err
	}

	// find open order
	orderID, err := s.Repo.FindOpenOrder(ctx, cid)
	if err != nil {
		return 0, errors.New("no open cart")
	}

	// get cart items (cart repo returns items + total)
	items, total, err := s.Repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, errors.New("cart is empty")
	}

	// build gameIDs slice and map gameid->title for error messages
	gameIDs := make([]int64, 0, len(items))
	gameTitles := make(map[int64]string, len(items))
	for _, it := range items {
		gameIDs = append(gameIDs, it.GameID)
		gameTitles[it.GameID] = it.Title
	}

	// check ownership: if any owned -> reject entire checkout (Option A)
	ownedGameID, err := s.CustomerGamesRepo.ExistsAnyOwned(ctx, cid, gameIDs)
	if err != nil {
		return 0, fmt.Errorf("ownership check failed: %w", err)
	}
	if ownedGameID != 0 {
		title := gameTitles[ownedGameID]
		if title == "" {
			// fallback to fetching title
			if t, _, terr := s.Repo.GetGameInfo(ctx, ownedGameID); terr == nil {
				title = t
			} else {
				title = "owned_game"
			}
		}
		return 0, fmt.Errorf("checkout rejected: already own game '%s' (id=%d)", title, ownedGameID)
	}

	// Begin transaction using cart repo's DB
	tx, err := s.Repo.DB.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1) finalize order (update totalprice and orderdate) using tx method
	if err := s.Repo.CheckoutOrderTx(ctx, tx, orderID, total); err != nil {
		return 0, fmt.Errorf("finalize order: %w", err)
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	// Return finalized orderID
	return orderID, nil
}
