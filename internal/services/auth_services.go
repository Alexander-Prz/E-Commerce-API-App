package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"os"
	"strings"
	"time"

	"GameStoreAPI/internal/model"
	"GameStoreAPI/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLen = 8
)

type AuthService struct {
	Auths          *repository.AuthRepository
	Customer       *repository.CustomerRepository
	EmailValidator EmailValidator

	Mailer     EmailSender
	VerifyRepo *repository.EmailVerificationRepository
}

func NewAuthService(
	authRepo *repository.AuthRepository,
	customerRepo *repository.CustomerRepository,
	emailValidator EmailValidator,
	mailer EmailSender,
	verifyRepo *repository.EmailVerificationRepository,
) *AuthService {
	return &AuthService{
		Auths:          authRepo,
		Customer:       customerRepo,
		EmailValidator: emailValidator,
		Mailer:         mailer,
		VerifyRepo:     verifyRepo,
	}
}

func (s *AuthService) validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email is required")
	}

	// 1️⃣ Local syntax validation (cheap)
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("invalid email format")
	}

	// 2️⃣ External validation (Abstract API)
	if s.EmailValidator != nil {
		if err := s.EmailValidator.Validate(
			context.Background(),
			email,
		); err != nil {
			return err
		}
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

	exists, err := s.Auths.EmailExists(ctx, email)
	if err != nil {
		return 0, err
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	authID, err := s.Auths.CreateUser(ctx, email, string(hash), "user")
	if err != nil {
		return 0, err
	}

	if _, err := s.Customer.Create(ctx, authID, email); err != nil {
		return authID, err
	}

	if exists {
		return 0, errors.New("email already registered")
	}

	// Email verification
	token := uuid.NewString()
	exp := time.Now().Add(24 * time.Hour)

	_ = s.VerifyRepo.Create(ctx, authID, token, exp)

	verifyURL := os.Getenv("APP_BASE_URL") +
		"/api/auth/verify-email?token=" + token

	// Non-blocking email send
	//go s.Mailer.SendVerificationEmail(context.Background(), email, verifyURL)
	if err := s.Mailer.SendVerificationEmail(
		context.Background(),
		email,
		verifyURL,
	); err != nil {
		log.Println("EMAIL SEND FAILED:", err)
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
	exists, err := s.Auths.EmailExists(ctx, email)
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
	return s.Auths.CreateUser(ctx, email, string(hash), role)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
)

// Login authenticates using email + password and returns the user (without passwordhash).
func (s *AuthService) Login(ctx context.Context, email, password string) (*model.Auth, error) {
	u, err := s.Auths.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !u.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(u.PasswordHash),
		[]byte(password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	u.PasswordHash = ""
	return u, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	authID, err := s.VerifyRepo.GetAuthID(ctx, token)
	if err != nil {
		return err
	}

	if err := s.Auths.SetEmailVerified(ctx, authID); err != nil {
		return err
	}

	_ = s.VerifyRepo.Delete(ctx, token)
	return nil
}
