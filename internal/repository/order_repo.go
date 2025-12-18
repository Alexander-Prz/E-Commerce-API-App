package repository

import (
	"context"
	"errors"
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

// CreateOrderFromCart creates a pending order from customer's cart
func (r *OrderRepository) CreateOrderFromCart(ctx context.Context, customerID int64) (int64, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Load cart items
	qCart := `
		SELECT gameid, quantity, price
		FROM cartitems
		WHERE customerid=$1 AND deleted_at IS NULL
	`
	rows, err := tx.Query(ctx, qCart, customerID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type item struct {
		gameID int64
		qty    int
		price  int64
	}

	var items []item
	var total int64

	for rows.Next() {
		var it item
		if err := rows.Scan(&it.gameID, &it.qty, &it.price); err != nil {
			return 0, err
		}
		items = append(items, it)
		total += int64(it.qty) * it.price
	}

	if len(items) == 0 {
		return 0, errors.New("cart is empty")
	}

	// Create order
	var orderID int64
	qOrder := `
		INSERT INTO orders (customerid, orderdate, totalprice)
		VALUES ($1, NOW(), $2)
		RETURNING orderid
	`
	if err := tx.QueryRow(ctx, qOrder, customerID, total).Scan(&orderID); err != nil {
		return 0, err
	}

	// Create order items
	qItem := `
		INSERT INTO orderitems (orderid, gameid, quantity, priceatpurchase)
		VALUES ($1, $2, $3, $4)
	`
	for _, it := range items {
		if _, err := tx.Exec(ctx, qItem,
			orderID, it.gameID, it.qty, it.price,
		); err != nil {
			return 0, err
		}
	}

	// Clear cart
	qClear := `
		UPDATE cartitems
		SET deleted_at=NOW()
		WHERE customerid=$1 AND deleted_at IS NULL
	`
	if _, err := tx.Exec(ctx, qClear, customerID); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return orderID, nil
}
