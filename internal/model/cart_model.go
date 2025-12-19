package model

import "time"

// Order represents an entry in the orders table
type Order struct {
	OrderID     int64      `json:"orderid"`
	CustomerID  int64      `json:"customerid"`
	OrderStatus string     `json:"orderstatus"`
	OrderDate   *time.Time `json:"orderdate,omitempty"`
	TotalPrice  *float64   `json:"totalprice,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// OrderItem represents a row in the orderitems table
type OrderItem struct {
	OrderItemID     int64      `json:"orderitemid"`
	OrderID         int64      `json:"orderid"`
	GameID          int64      `json:"gameid"`
	Quantity        int        `json:"quantity"`
	PriceAtPurchase float64    `json:"priceatpurchase"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

// CartItem is what the API exposes (joined with games.title)
type CartItem struct {
	OrderItemID     int64   `json:"orderitemid"`
	GameID          int64   `json:"gameid"`
	Title           string  `json:"title"`
	Quantity        int     `json:"quantity"`
	PriceAtPurchase float64 `json:"priceatpurchase"`
	Subtotal        float64 `json:"subtotal"`
}

// CartResponse is returned when calling GET /api/cart
type CartResponse struct {
	Items []CartItem `json:"items"`
	Total float64    `json:"total"`
}
