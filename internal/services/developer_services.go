package services

import (
	"context"
	"errors"
	"strings"

	"GameStoreAPI/internal/model"
	"GameStoreAPI/internal/repository"
)

type DeveloperService struct {
	Repo *repository.DeveloperRepository
}

func NewDeveloperService(r *repository.DeveloperRepository) *DeveloperService {
	return &DeveloperService{Repo: r}
}

func (s *DeveloperService) CreateDeveloper(ctx context.Context, name string, authID *int64) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("developer name is required")
	}
	exists, err := s.Repo.NameExists(ctx, name)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, errors.New("developer with this name already exists")
	}
	return s.Repo.CreateDeveloper(ctx, name, authID)
}

func (s *DeveloperService) GetDeveloper(ctx context.Context, id int64) (*model.Developer, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *DeveloperService) ListDevelopers(ctx context.Context) ([]model.Developer, error) {
	return s.Repo.GetAll(ctx)
}

func (s *DeveloperService) ListDevelopersAdmin(ctx context.Context) ([]model.Developer, error) {
	return s.Repo.GetAllAdmin(ctx)
}

func (s *DeveloperService) UpdateDeveloper(ctx context.Context, id int64, name string, authID *int64) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("developer name is required")
	}
	// optional: check uniqueness (skip if same as current)
	// For simplicity, check name exists and is not the same id is left to DB constraints or frontend
	return s.Repo.UpdateDeveloper(ctx, id, name, authID)
}

func (s *DeveloperService) DeleteDeveloper(ctx context.Context, id int64) error {
	return s.Repo.DeleteDeveloper(ctx, id)
}
