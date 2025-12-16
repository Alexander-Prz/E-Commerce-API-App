package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"GameStoreAPI/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerGamesRepository struct {
	DB *pgxpool.Pool
}

func NewCustomerGamesRepository(db *pgxpool.Pool) *CustomerGamesRepository {
	return &CustomerGamesRepository{DB: db}
}

// ExistsAnyOwned returns the first gameid from gameIDs that the customer already owns.
// Returns 0 and nil if none are owned.
func (r *CustomerGamesRepository) ExistsAnyOwned(ctx context.Context, customerID int64, gameIDs []int64) (int64, error) {
	if len(gameIDs) == 0 {
		return 0, nil
	}
	// Use Postgres ANY with array
	q := `SELECT gameid FROM customer_games
		WHERE customerid = $1 AND gameid = ANY($2)
		LIMIT 1`
	var gid int64
	err := r.DB.QueryRow(ctx, q, customerID, gameIDs).Scan(&gid)

	if err != nil {
		// no rows -> not owned
		// pgx returns pgx.ErrNoRows
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return gid, nil
}

// CreateCustomerGamesTx inserts ownership records inside the provided tx.
// It inserts multiple rows with a single INSERT statement. Caller should ensure duplicates are handled by UNIQUE constraint.
func (r *CustomerGamesRepository) CreateCustomerGamesTx(ctx context.Context, tx pgx.Tx, customerID int64, gameIDs []int64) error {
	if len(gameIDs) == 0 {
		return nil
	}
	// Build batch INSERT
	var sb strings.Builder
	args := make([]interface{}, 0, len(gameIDs)*3)
	sb.WriteString("INSERT INTO customer_games (customerid, gameid, purchased_at) VALUES ")
	for i, gid := range gameIDs {
		if i > 0 {
			sb.WriteString(",")
		}
		// placeholders: ($1,$2,$3), ($4,$5,$6), ...
		pi := i*3 + 1
		sb.WriteString(fmt.Sprintf("($%d,$%d,$%d)", pi, pi+1, pi+2))
		args = append(args, customerID, gid, time.Now())
	}
	// Use ON CONFLICT DO NOTHING to avoid failing on duplicates (though we already check ownership before checkout)
	_, err := tx.Exec(ctx, sb.String()+" ON CONFLICT (customerid, gameid) DO NOTHING", args...)
	return err
}

// NOTE: If you want non-TX helpers later, add them here.

// InsertPurchased inserts only NEW ownership records
func (r *CustomerGamesRepository) InsertPurchased(ctx context.Context, customerID int64, gameIDs []int64) error {
	if len(gameIDs) == 0 {
		return nil
	}

	query := `
        INSERT INTO customer_games (customerid, gameid)
        VALUES ($1, UNNEST($2::bigint[]))
        ON CONFLICT DO NOTHING
    `
	_, err := r.DB.Exec(ctx, query, customerID, gameIDs)
	return err
}

// ListOwnedGames returns all games the customer owns
func (r *CustomerGamesRepository) ListOwnedGames(ctx context.Context, customerID int64) ([]model.Game, error) {
	query := `
        SELECT g.gameid, g.title, g.price, g.releasedate, g.developerid
        FROM customer_games cg
        JOIN games g ON g.gameid = cg.gameid
        WHERE cg.customerid = $1 AND g.deleted_at IS NULL
    `
	rows, err := r.DB.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Game
	for rows.Next() {
		var g model.Game
		if err := rows.Scan(
			&g.GameID,
			&g.Title,
			&g.Price,
			&g.ReleaseDate,
			&g.DeveloperID,
		); err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, nil
}

// ListOrders returns all completed order headers
func (r *CustomerGamesRepository) ListOrders(ctx context.Context, customerID int64) ([]model.Order, error) {
	query := `
        SELECT orderid, customerid, orderdate, totalprice, created_at
        FROM orders
        WHERE customerid=$1 AND totalprice IS NOT NULL
        ORDER BY created_at DESC
    `
	rows, err := r.DB.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.OrderID, &o.CustomerID, &o.OrderDate, &o.TotalPrice, &o.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, nil
}

// GetOrderDetails returns one order with items
func (r *CustomerGamesRepository) GetOrderDetails(ctx context.Context, customerID, orderID int64) (*model.Order, []model.OrderItem, error) {
	var o model.Order
	q1 := `
        SELECT orderid, customerid, orderdate, totalprice, created_at
        FROM orders
        WHERE orderid=$1 AND customerid=$2
    `
	if err := r.DB.QueryRow(ctx, q1, orderID, customerID).Scan(
		&o.OrderID, &o.CustomerID, &o.OrderDate, &o.TotalPrice, &o.CreatedAt,
	); err != nil {
		return nil, nil, errors.New("order not found")
	}

	q2 := `
        SELECT orderitemid, gameid, quantity, priceatpurchase, created_at
        FROM orderitems
        WHERE orderid=$1 AND deleted_at IS NULL
    `
	rows, err := r.DB.Query(ctx, q2, orderID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var it model.OrderItem
		if err := rows.Scan(&it.OrderItemID, &it.GameID, &it.Quantity, &it.PriceAtPurchase, &it.CreatedAt); err != nil {
			return nil, nil, err
		}
		items = append(items, it)
	}

	return &o, items, nil
}

// ListAllOrders returns all completed orders across all users
func (r *CustomerGamesRepository) ListAllOrders(ctx context.Context) ([]model.Order, error) {
	query := `
        SELECT orderid, customerid, orderdate, totalprice, created_at
        FROM orders
        WHERE totalprice IS NOT NULL
        ORDER BY created_at DESC
    `
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.OrderID, &o.CustomerID, &o.OrderDate, &o.TotalPrice, &o.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, nil
}

// GetOrderDetailsAdmin returns order and items without checking customer ownership
func (r *CustomerGamesRepository) GetOrderDetailsAdmin(ctx context.Context, orderID int64) (*model.Order, []model.OrderItem, error) {
	var o model.Order
	q1 := `
        SELECT orderid, customerid, orderdate, totalprice, created_at
        FROM orders
        WHERE orderid=$1
    `
	if err := r.DB.QueryRow(ctx, q1, orderID).Scan(
		&o.OrderID, &o.CustomerID, &o.OrderDate, &o.TotalPrice, &o.CreatedAt,
	); err != nil {
		return nil, nil, errors.New("order not found")
	}

	q2 := `
        SELECT orderitemid, gameid, quantity, priceatpurchase, created_at
        FROM orderitems
        WHERE orderid=$1 AND deleted_at IS NULL
    `
	rows, err := r.DB.Query(ctx, q2, orderID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var it model.OrderItem
		if err := rows.Scan(&it.OrderItemID, &it.GameID, &it.Quantity, &it.PriceAtPurchase, &it.CreatedAt); err != nil {
			return nil, nil, err
		}
		items = append(items, it)
	}

	return &o, items, nil
}
