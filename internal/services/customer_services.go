package services

import (
	"context"
	"errors"

	"GameStoreAPI/internal/model"
	"GameStoreAPI/internal/repository"
)

type CustomerService struct {
	Customers *repository.CustomerRepository
	Users     *repository.AuthRepository
}

func NewCustomerService(cr *repository.CustomerRepository, ur *repository.AuthRepository) *CustomerService {
	return &CustomerService{Customers: cr, Users: ur}
}

// CreateForNewUser creates a customer row for a newly-registered public user.
// This is called after a successful RegisterPublic in AuthService.
func (s *CustomerService) CreateForNewUser(ctx context.Context, authID int64, email string) (int64, error) {
	// only create if the user exists and has role "user"
	u, err := s.Users.GetByID(ctx, authID)
	if err != nil {
		return 0, errors.New("user not found")
	}
	if u.Role != "user" {
		return 0, errors.New("customer created only for role=user")
	}
	return s.Customers.Create(ctx, authID, email)
}

func (s *CustomerService) GetByAuthID(ctx context.Context, authID int64) (*model.Customer, error) {
	return s.Customers.GetByAuthID(ctx, authID)
}

func (s *CustomerService) UpdateSelf(ctx context.Context, customerID int64, fullname, address, phone *string) error {
	return s.Customers.Update(ctx, customerID, fullname, address, phone)
}

func (s *CustomerService) BanUser(ctx context.Context, authID int64) error {
	return s.Users.BanUser(ctx, authID)
}
