package services

import (
	"context"
	"errors"
	"strings"

	"GameStoreAPI/internal/model"
	"GameStoreAPI/internal/repository"
)

type GameService struct {
	Repo          *repository.GameRepository
	DeveloperRepo *repository.DeveloperRepository
}

func NewGameService(r *repository.GameRepository, dr *repository.DeveloperRepository) *GameService {
	return &GameService{Repo: r, DeveloperRepo: dr}
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
