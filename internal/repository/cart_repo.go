package repository

import (
	"context"
	"errors"
	"time"

	"GameStoreAPI/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CartRepository struct {
	DB *pgxpool.Pool
}

func NewCartRepository(db *pgxpool.Pool) *CartRepository {
	return &CartRepository{DB: db}
}

// getCustomerID returns customerid for a given authid
func (r *CartRepository) GetCustomerID(ctx context.Context, authID int64) (int64, error) {
	var cid int64
	query := `SELECT customerid FROM customers WHERE authid=$1 AND deleted_at IS NULL`
	if err := r.DB.QueryRow(ctx, query, authID).Scan(&cid); err != nil {
		return 0, errors.New("customer not found")
	}
	return cid, nil
}

// findOpenOrder finds an order for customer where totalprice IS NULL and deleted_at IS NULL
func (r *CartRepository) FindOpenOrder(ctx context.Context, customerID int64) (int64, error) {
	var orderID int64
	query := `SELECT orderid FROM orders WHERE customerid=$1 AND totalprice IS NULL AND deleted_at IS NULL LIMIT 1`
	if err := r.DB.QueryRow(ctx, query, customerID).Scan(&orderID); err != nil {
		return 0, err
	}
	return orderID, nil
}

// createOpenOrder creates a new order with totalprice = NULL and returns orderid
func (r *CartRepository) CreateOpenOrder(ctx context.Context, customerID int64) (int64, error) {
	var orderID int64
	query := `INSERT INTO orders (customerid, orderdate, totalprice, created_at) VALUES ($1, $2, NULL, $3) RETURNING orderid`
	if err := r.DB.QueryRow(ctx, query, customerID, time.Now(), time.Now()).Scan(&orderID); err != nil {
		return 0, err
	}
	return orderID, nil
}

// getGamePrice gets the current games.price (numeric) and title
func (r *CartRepository) GetGameInfo(ctx context.Context, gameID int64) (title string, price float64, err error) {
	query := `SELECT title, price FROM games WHERE gameid=$1 AND deleted_at IS NULL`
	if err := r.DB.QueryRow(ctx, query, gameID).Scan(&title, &price); err != nil {
		return "", 0, errors.New("game not found")
	}
	return title, price, nil
}

// addOrIncrementOrderItem inserts or increments an item quantity for an order
func (r *CartRepository) AddOrIncrementOrderItem(ctx context.Context, orderID, gameID int64, qty int, priceAtPurchase float64) error {
	// If orderitem exists, update quantity; else insert
	query := `
		INSERT INTO orderitems (orderid, gameid, quantity, priceatpurchase, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (orderid, gameid)
		DO UPDATE SET quantity = orderitems.quantity + EXCLUDED.quantity
	`
	_, err := r.DB.Exec(ctx, query, orderID, gameID, qty, priceAtPurchase, time.Now())
	return err
}

// setOrderItemQuantity sets exact quantity for an orderitem
func (r *CartRepository) SetOrderItemQuantity(ctx context.Context, orderID, gameID int64, qty int) error {
	query := `UPDATE orderitems SET quantity=$1 WHERE orderid=$2 AND gameid=$3 AND deleted_at IS NULL`
	tag, err := r.DB.Exec(ctx, query, qty, orderID, gameID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("cart item not found")
	}
	return nil
}

// removeOrderItem removes a specific order item
func (r *CartRepository) RemoveOrderItem(ctx context.Context, orderID, gameID int64) error {
	query := `DELETE FROM orderitems WHERE orderid=$1 AND gameid=$2`
	_, err := r.DB.Exec(ctx, query, orderID, gameID)
	return err
}

// clearOrderItems clears all items for an order
func (r *CartRepository) ClearOrderItems(ctx context.Context, orderID int64) error {
	query := `DELETE FROM orderitems WHERE orderid=$1`
	_, err := r.DB.Exec(ctx, query, orderID)
	return err
}

// getOrderItems returns cart items for an order, with priceatpurchase and title
func (r *CartRepository) GetOrderItems(ctx context.Context, orderID int64) ([]model.CartItem, float64, error) {
	query := `
		SELECT oi.orderitemid, oi.gameid, g.title, oi.quantity, oi.priceatpurchase
		FROM orderitems oi
		JOIN games g ON g.gameid = oi.gameid
		WHERE oi.orderid=$1 AND oi.deleted_at IS NULL
	`
	rows, err := r.DB.Query(ctx, query, orderID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.CartItem
	var total float64
	for rows.Next() {
		var it model.CartItem
		if err := rows.Scan(&it.OrderItemID, &it.GameID, &it.Title, &it.Quantity, &it.PriceAtPurchase); err != nil {
			return nil, 0, err
		}
		it.Subtotal = it.PriceAtPurchase * float64(it.Quantity)
		items = append(items, it)
		total += it.Subtotal
	}
	return items, total, nil
}

// checkoutOrder sets totalprice on order to finalize it
func (r *CartRepository) CheckoutOrder(ctx context.Context, orderID int64, total float64) error {
	query := `UPDATE orders SET totalprice=$1, created_at=created_at WHERE orderid=$2` // keep created_at, orderdate default used by DB
	_, err := r.DB.Exec(ctx, query, total, orderID)
	return err
}

// CheckoutOrderTx updates order totalprice and orderdate inside a transaction.
func (r *CartRepository) CheckoutOrderTx(ctx context.Context, tx pgx.Tx, orderID int64, total float64) error {
	// set totalprice and update orderdate to now
	query := `UPDATE orders SET totalprice=$1, orderdate=$2 WHERE orderid=$3`
	_, err := tx.Exec(ctx, query, total, time.Now(), orderID)
	return err
}
