package services

import (
	"context"
	"errors"

	"GameStoreAPI/internal/repository"
)

type GameGenreService struct {
	Repo      *repository.GameGenreRepository
	GameRepo  *repository.GameRepository
	GenreRepo *repository.GenreRepository
}

func NewGameGenreService(r *repository.GameGenreRepository, gr *repository.GameRepository, ge *repository.GenreRepository) *GameGenreService {
	return &GameGenreService{Repo: r, GameRepo: gr, GenreRepo: ge}
}

func (s *GameGenreService) Add(ctx context.Context, gameID, genreID int64) error {
	// ensure game exists
	if _, err := s.GameRepo.GetByID(ctx, gameID); err != nil {
		return errors.New("game not found")
	}
	// ensure genre exists
	if _, err := s.GenreRepo.GetByID(ctx, genreID); err != nil {
		return errors.New("genre not found")
	}
	return s.Repo.AddGenreToGame(ctx, gameID, genreID)
}

func (s *GameGenreService) Remove(ctx context.Context, gameID, genreID int64) error {
	// ensure game exists
	if _, err := s.GameRepo.GetByID(ctx, gameID); err != nil {
		return errors.New("game not found")
	}
	// ensure genre exists
	if _, err := s.GenreRepo.GetByID(ctx, genreID); err != nil {
		return errors.New("genre not found")
	}
	return s.Repo.RemoveGenreFromGame(ctx, gameID, genreID)
}

func (s *GameGenreService) ListGenres(ctx context.Context, gameID int64) ([]int64, error) {
	return s.Repo.GetGenresByGame(ctx, gameID)
}
