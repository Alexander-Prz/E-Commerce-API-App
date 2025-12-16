package repository

import (
	"context"
	"time"

	"GameStoreAPI/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	DB *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{DB: db}
}

// GetOrdersByCustomer returns orders for a given customerid (completed orders where totalprice IS NOT NULL).
func (r *OrderRepository) GetOrdersByCustomer(ctx context.Context, customerID int64) ([]model.Order, error) {
	query := `SELECT orderid, customerid, totalprice, orderdate, created_at, deleted_at FROM orders WHERE customerid=$1 AND totalprice IS NOT NULL ORDER BY orderid DESC`
	rows, err := r.DB.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Order
	for rows.Next() {
		var o model.Order
		var tp *float64
		var od *time.Time
		if err := rows.Scan(&o.OrderID, &o.CustomerID, &tp, &od, &o.CreatedAt, &o.DeletedAt); err != nil {
			return nil, err
		}
		o.TotalPrice = tp
		o.OrderDate = od
		out = append(out, o)
	}
	return out, nil
}

// GetOrderByID returns the order row for the given orderid
func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID int64) (*model.Order, error) {
	query := `SELECT orderid, customerid, totalprice, orderdate, created_at, deleted_at FROM orders WHERE orderid=$1`
	var o model.Order
	var tp *float64
	var od *time.Time
	if err := r.DB.QueryRow(ctx, query, orderID).Scan(&o.OrderID, &o.CustomerID, &tp, &od, &o.CreatedAt, &o.DeletedAt); err != nil {
		return nil, err
	}
	o.TotalPrice = tp
	o.OrderDate = od
	return &o, nil
}
