package services

import (
	"context"
	"errors"
	"strings"

	"GameStoreAPI/internal/model"
	"GameStoreAPI/internal/repository"
)

type GenreService struct {
	Repo *repository.GenreRepository
}

func NewGenreService(r *repository.GenreRepository) *GenreService {
	return &GenreService{Repo: r}
}

func (s *GenreService) Create(ctx context.Context, name string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("genre name is required")
	}
	exists, err := s.Repo.ExistsByName(ctx, name)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, errors.New("genre already exists")
	}
	return s.Repo.Create(ctx, name)
}

func (s *GenreService) Get(ctx context.Context, id int64) (*model.Genre, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *GenreService) List(ctx context.Context) ([]model.Genre, error) {
	return s.Repo.List(ctx)
}

func (s *GenreService) Update(ctx context.Context, id int64, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("genre name is required")
	}
	return s.Repo.Update(ctx, id, name)
}

func (s *GenreService) Delete(ctx context.Context, id int64) error {
	return s.Repo.Delete(ctx, id)
}
