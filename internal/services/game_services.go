package services

import (
	"context"
	"errors"
	"strings"

	"GameStoreAPI/internal/model"
	"GameStoreAPI/internal/repository"
)

type GameService struct {
	Repo             *repository.GameRepository
	DeveloperRepo    *repository.DeveloperRepository
	CustomerRepo     *repository.CustomerRepository
	CustomerGameRepo *repository.CustomerGamesRepository
}

func NewGameService(r *repository.GameRepository, dr *repository.DeveloperRepository, cr *repository.CustomerRepository,
	cgr *repository.CustomerGamesRepository) *GameService {
	return &GameService{Repo: r, DeveloperRepo: dr, CustomerRepo: cr, CustomerGameRepo: cgr}
}

func (s *GameService) CreateGame(ctx context.Context, g *model.Game) (int64, error) {
	// validate title and developer existence
	g.Title = strings.TrimSpace(g.Title)
	if g.Title == "" {
		return 0, errors.New("title is required")
	}
	if g.Price < 0 {
		return 0, errors.New("price must be >= 0")
	}
	ok, err := s.Repo.ExistsByDeveloperID(ctx, g.DeveloperID)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, errors.New("developer not found")
	}
	return s.Repo.CreateGame(ctx, g)
}

func (s *GameService) GetGame(ctx context.Context, id int64) (*model.Game, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *GameService) ListGames(ctx context.Context, limit, offset int) ([]model.Game, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.Repo.List(ctx, limit, offset)
}

func (s *GameService) ListGamesByDeveloper(ctx context.Context, developerID int64, limit, offset int) ([]model.Game, error) {
	// simple validation
	if developerID <= 0 {
		return nil, errors.New("invalid developer id")
	}
	return s.Repo.ListByDeveloper(ctx, developerID, limit, offset)
}

func (s *GameService) UpdateGame(ctx context.Context, g *model.Game) error {
	g.Title = strings.TrimSpace(g.Title)
	if g.Title == "" {
		return errors.New("title is required")
	}
	if g.Price < 0 {
		return errors.New("price must be >= 0")
	}
	ok, err := s.Repo.ExistsByDeveloperID(ctx, g.DeveloperID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("developer not found")
	}
	return s.Repo.UpdateGame(ctx, g)
}

func (s *GameService) DeleteGame(ctx context.Context, id int64) error {
	return s.Repo.DeleteGame(ctx, id)
}

func (s *GameService) ListGamesWithOwnership(
	ctx context.Context,
	authID *int64,
	role *string,
	limit, offset int,
) ([]model.Game, map[int64]bool, error) {

	games, err := s.ListGames(ctx, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	// no auth OR not user → no ownership check
	if authID == nil || role == nil || *role != "user" {
		return games, nil, nil
	}

	customer, err := s.CustomerRepo.GetByAuthID(ctx, *authID)
	if err != nil {
		return games, nil, nil // user exists but no customer row yet
	}

	ownedMap, err := s.CustomerGameRepo.ListOwnedGameIDs(ctx, customer.CustomerID)
	if err != nil {
		return nil, nil, err
	}

	return games, ownedMap, nil
}

func (s *GameService) GetGameWithOwnership(
	ctx context.Context,
	gameID int64,
	authID *int64,
	role *string,
) (*model.Game, bool, error) {

	game, err := s.GetGame(ctx, gameID)
	if err != nil {
		return nil, false, err
	}

	if authID == nil || role == nil || *role != "user" {
		return game, false, nil
	}

	customer, err := s.CustomerRepo.GetByAuthID(ctx, *authID)
	if err != nil {
		return game, false, nil
	}

	ownedMap, err := s.CustomerGameRepo.ListOwnedGameIDs(ctx, customer.CustomerID)
	if err != nil {
		return nil, false, err
	}

	return game, ownedMap[gameID], nil
}
