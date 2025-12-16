package repository

import (
	"context"
	"errors"
	"time"

	"GameStoreAPI/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerRepository struct {
	DB *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{DB: db}
}

// Create creates a customer row (used only during public registration)
func (r *CustomerRepository) Create(ctx context.Context, authID int64, email string) (int64, error) {
	var id int64
	query := `
		INSERT INTO customers (authid, email, created_at)
		VALUES ($1, $2, $3)
		RETURNING customerid
	`
	if err := r.DB.QueryRow(ctx, query, authID, email, time.Now()).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// GetByAuthID returns a customer by authid
func (r *CustomerRepository) GetByAuthID(ctx context.Context, authID int64) (*model.Customer, error) {
	var c model.Customer
	query := `SELECT customerid, authid, username, fullname, email, address, phone, created_at, deleted_at FROM customers WHERE authid=$1 AND deleted_at IS NULL`
	if err := r.DB.QueryRow(ctx, query, authID).Scan(&c.CustomerID, &c.AuthID, &c.Username, &c.Fullname, &c.Email, &c.Address, &c.Phone, &c.CreatedAt, &c.DeletedAt); err != nil {
		return nil, errors.New("customer not found")
	}
	return &c, nil
}

// GetByID returns a customer by customerid (internal use)
func (r *CustomerRepository) GetByID(ctx context.Context, id int64) (*model.Customer, error) {
	var c model.Customer
	query := `SELECT customerid, authid, username, fullname, email, address, phone, created_at, deleted_at FROM customers WHERE customerid=$1`
	if err := r.DB.QueryRow(ctx, query, id).Scan(&c.CustomerID, &c.AuthID, &c.Username, &c.Fullname, &c.Email, &c.Address, &c.Phone, &c.CreatedAt, &c.DeletedAt); err != nil {
		return nil, errors.New("customer not found")
	}
	return &c, nil
}

// Update allows a user to update their own customer record
func (r *CustomerRepository) Update(ctx context.Context, id int64, fullname, address, phone *string) error {
	query := `UPDATE customers SET fullname=$1, address=$2, phone=$3 WHERE customerid=$4 AND deleted_at IS NULL`
	tag, err := r.DB.Exec(ctx, query, fullname, address, phone, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("customer not found or deleted")
	}
	return nil
}

// ListAll returns all customers (admin use). Note: Personal fields are returned here;
// admin handlers should redact if privacy requires.
func (r *CustomerRepository) ListAll(ctx context.Context) ([]model.Customer, error) {
	query := `SELECT customerid, authid, username, fullname, email, address, phone, created_at, deleted_at FROM customers ORDER BY customerid`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Customer
	for rows.Next() {
		var c model.Customer
		if err := rows.Scan(&c.CustomerID, &c.AuthID, &c.Username, &c.Fullname, &c.Email, &c.Address, &c.Phone, &c.CreatedAt, &c.DeletedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}
