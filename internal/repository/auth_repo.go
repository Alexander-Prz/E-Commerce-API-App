package repository

import (
	"context"
	"errors"
	"time"

	"GameStoreAPI/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	DB *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{DB: db}
}

type MinimalUser struct {
	AuthID     int64     `json:"authid"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	CustomerID int64     `json:"customerid"`
	CreatedAt  time.Time `json:"created_at"`
	Banned     bool      `json:"banned"`
}

// CreateUser inserts a new user and returns the created authid
func (r *AuthRepository) CreateUser(ctx context.Context, email, passwordhash, role string) (int64, error) {
	var id int64
	query := `INSERT INTO userauth (email, passwordhash, role, created_at, email_verified) VALUES ($1, $2, $3, $4, true) RETURNING authid`
	if err := r.DB.QueryRow(ctx, query, email, passwordhash, role, time.Now()).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *AuthRepository) GetByEmail(ctx context.Context, email string) (*model.Auth, error) {
	var u model.Auth
	query := `SELECT authid, email, passwordhash, role, email_verified, created_at, deleted_at
			FROM userauth
			WHERE email=$1`
	if err := r.DB.QueryRow(ctx, query, email).Scan(&u.AuthID, &u.Email, &u.PasswordHash, &u.Role, &u.EmailVerified, &u.CreatedAt, &u.DeletedAt); err != nil {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

func (r *AuthRepository) GetByID(ctx context.Context, id int64) (*model.Auth, error) {
	var u model.Auth
	query := `SELECT authid, email, role, created_at, deleted_at FROM userauth WHERE authid=$1`
	if err := r.DB.QueryRow(ctx, query, id).Scan(&u.AuthID, &u.Email, &u.Role, &u.CreatedAt, &u.DeletedAt); err != nil {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

func (r *AuthRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM userauth WHERE email=$1)`
	if err := r.DB.QueryRow(ctx, query, email).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// LIST ONLY USERS (role=user)
func (r *AuthRepository) ListUsersOnly(ctx context.Context) ([]MinimalUser, error) {
	q := `
        SELECT u.authid, u.email, u.role, u.created_at,
               COALESCE(c.customerid, 0) AS customerid,
               (u.deleted_at IS NOT NULL) AS banned
        FROM userauth u
        LEFT JOIN customers c ON c.authid = u.authid
        WHERE u.role = 'user'
        ORDER BY u.authid;
    `
	rows, err := r.DB.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []MinimalUser{}
	for rows.Next() {
		var m MinimalUser
		if err := rows.Scan(&m.AuthID, &m.Email, &m.Role, &m.CreatedAt, &m.CustomerID, &m.Banned); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

// GET one user (role=user only)
func (r *AuthRepository) GetUserOnlyByID(ctx context.Context, authID int64) (*MinimalUser, error) {
	q := `
        SELECT u.authid, u.email, u.role, u.created_at,
               COALESCE(c.customerid, 0) AS customerid,
               (u.deleted_at IS NOT NULL) AS banned
        FROM userauth u
        LEFT JOIN customers c ON c.authid = u.authid
        WHERE u.role = 'user' AND u.authid = $1;
    `

	var m MinimalUser
	err := r.DB.QueryRow(ctx, q, authID).
		Scan(&m.AuthID, &m.Email, &m.Role, &m.CreatedAt, &m.CustomerID, &m.Banned)

	if err != nil {
		return nil, err
	}
	return &m, nil
}

// BanUser soft-deletes a user (sets deleted_at)
func (r *AuthRepository) BanUser(ctx context.Context, authID int64) error {
	query := `UPDATE userauth SET deleted_at=$1 WHERE authid=$2 AND deleted_at IS NULL`
	tag, err := r.DB.Exec(ctx, query, time.Now(), authID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("user not found or already banned")
	}
	return nil
}

func (r *AuthRepository) UnBanUser(ctx context.Context, authID int64) error {
	query := `UPDATE userauth SET deleted_at=NULL WHERE authid=$1 AND deleted_at IS NOT NULL`
	tag, err := r.DB.Exec(ctx, query, authID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("user not found or already unbanned")
	}
	return nil
}

func (r *AuthRepository) SetEmailVerified(ctx context.Context, authID int64) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE userauth
		SET email_verified = TRUE
		WHERE authid = $1
	`, authID)
	return err
}
