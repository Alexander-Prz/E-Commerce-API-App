package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"GameStoreAPI/internal/model"
	"GameStoreAPI/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLen = 8
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

type AuthService struct {
	Users    *repository.AuthRepository
	Customer *repository.CustomerRepository // for auto-create
}

func NewAuthService(u *repository.AuthRepository, cr *repository.CustomerRepository) *AuthService {
	return &AuthService{Users: u, Customer: cr}
}

func (s *AuthService) validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

func (s *AuthService) validatePassword(pw string) error {
	if len(pw) < MinPasswordLen {
		return fmt.Errorf("password too short: must be at least %d characters", MinPasswordLen)
	}
	return nil
}

// RegisterPublic creates a user with role "user" AND creates the customer row.
func (s *AuthService) RegisterPublic(ctx context.Context, email, password string) (int64, error) {
	if err := s.validateEmail(email); err != nil {
		return 0, err
	}
	if err := s.validatePassword(password); err != nil {
		return 0, err
	}
	exists, err := s.Users.EmailExists(ctx, email)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, errors.New("email already registered")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	authID, err := s.Users.CreateUser(ctx, email, string(hash), "user")
	if err != nil {
		return 0, err
	}
	// create customer row
	if _, err := s.Customer.Create(ctx, authID, email); err != nil {
		// If creating customer fails, you might want to rollback user creation.
		// For now, return the authID and the error so caller can decide.
		return authID, err
	}
	return authID, nil
}

// RegisterByAdmin is still available but admin endpoints must ensure role != "user"
func (s *AuthService) RegisterByAdmin(ctx context.Context, email, password, role string) (int64, error) {
	if role == "" {
		return 0, errors.New("role required")
	}
	if role == "user" {
		return 0, errors.New("admins cannot create user accounts")
	}
	if err := s.validateEmail(email); err != nil {
		return 0, err
	}
	if err := s.validatePassword(password); err != nil {
		return 0, err
	}
	exists, err := s.Users.EmailExists(ctx, email)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, errors.New("email already registered")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	return s.Users.CreateUser(ctx, email, string(hash), role)
}

// Login authenticates using email + password and returns the user (without passwordhash).
func (s *AuthService) Login(ctx context.Context, email, password string) (*model.Auth, error) {
	u, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		// do not reveal whether email exists
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	// zero out password before returning
	u.PasswordHash = ""
	return u, nil
}
