package repository

import (
	"context"
	"errors"
	"time"

	"GameStoreAPI/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	DB                *pgxpool.Pool
	CustomerGamesRepo *CustomerGamesRepository
}

func NewOrderRepository(db *pgxpool.Pool, cgr *CustomerGamesRepository) *OrderRepository {
	return &OrderRepository{DB: db, CustomerGamesRepo: cgr}
}

// GetOrdersByCustomer returns orders for a given customerid (completed orders where totalprice IS NOT NULL).
func (r *OrderRepository) GetOrdersByCustomer(ctx context.Context, customerID int64) ([]model.Order, error) {
	query := `
		SELECT orderid, customerid, totalprice, orderstatus,
		       orderdate, created_at, deleted_at
		FROM orders
		WHERE customerid=$1
		  AND totalprice IS NOT NULL
		ORDER BY orderid DESC
	`

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

		if err := rows.Scan(
			&o.OrderID,
			&o.CustomerID,
			&tp,
			&o.OrderStatus,
			&od,
			&o.CreatedAt,
			&o.DeletedAt,
		); err != nil {
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
	query := `
		SELECT orderid, customerid, totalprice, orderstatus,
		       orderdate, created_at, deleted_at
		FROM orders
		WHERE orderid=$1
	`

	var o model.Order
	var tp *float64
	var od *time.Time

	if err := r.DB.QueryRow(ctx, query, orderID).Scan(
		&o.OrderID,
		&o.CustomerID,
		&tp,
		&o.OrderStatus,
		&od,
		&o.CreatedAt,
		&o.DeletedAt,
	); err != nil {
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

	// 1️⃣ Find active cart order
	var orderID int64
	qCartOrder := `
		SELECT orderid
		FROM orders
		WHERE customerid=$1
		  AND orderstatus='PendingPayment'
		  AND totalprice IS NULL
		  AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	if err := tx.QueryRow(ctx, qCartOrder, customerID).Scan(&orderID); err != nil {
		return 0, errors.New("cart is empty")
	}

	// 2️⃣ Load cart items
	qItems := `
		SELECT gameid, quantity, priceatpurchase
		FROM orderitems
		WHERE orderid=$1
	`
	rows, err := tx.Query(ctx, qItems, orderID)
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
	var gameIDs []int64
	var total int64

	for rows.Next() {
		var it item
		if err := rows.Scan(&it.gameID, &it.qty, &it.price); err != nil {
			return 0, err
		}
		items = append(items, it)
		gameIDs = append(gameIDs, it.gameID)
		total += int64(it.qty) * it.price
	}

	if len(items) == 0 {
		return 0, errors.New("cart is empty")
	}

	// 3️⃣ Prevent repurchasing owned games
	ownedGameID, err := r.CustomerGamesRepo.ExistsAnyOwned(ctx, customerID, gameIDs)
	if err != nil {
		return 0, err
	}
	if ownedGameID != 0 {
		return 0, errors.New("checkout contains already owned game")
	}

	// 4️⃣ Finalize order (cart → checkout)
	qFinalize := `
		UPDATE orders
		SET totalprice=$1,
		    orderdate=NOW()
		WHERE orderid=$2
		  AND orderstatus='PendingPayment'
	`
	if _, err := tx.Exec(ctx, qFinalize, total, orderID); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return orderID, nil
}

/* NON TRANSACTION
func (r *OrderRepository) MarkPaid(ctx context.Context, orderID int64) error {

	q := `
		UPDATE orders
		SET orderstatus='Paid'
		WHERE orderid=$1
		  AND orderstatus='PendingPayment'
	`
	_, err := r.DB.Exec(ctx, q, orderID)
	return err
}
*/

func (r *OrderRepository) MarkPaidTx(
	ctx context.Context,
	tx pgx.Tx,
	orderID int64,
) error {
	_, err := tx.Exec(ctx, `
		UPDATE orders
		SET orderstatus='Paid'
		WHERE orderid=$1 AND orderstatus='PendingPayment'
	`, orderID)
	return err
}

func (r *OrderRepository) MarkFailed(ctx context.Context, orderID int64) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE orders
		SET orderstatus='Failed'
		WHERE orderid=$1
		  AND orderstatus IN ('PendingPayment')
	`, orderID)
	return err
}

func (r *OrderRepository) GetGameIDsByOrderID(
	ctx context.Context,
	orderID int64,
) ([]int64, error) {

	rows, err := r.DB.Query(ctx,
		`SELECT gameid FROM orderitems WHERE orderid=$1`,
		orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gameIDs []int64
	for rows.Next() {
		var gid int64
		if err := rows.Scan(&gid); err != nil {
			return nil, err
		}
		gameIDs = append(gameIDs, gid)
	}

	return gameIDs, nil
}
