package model

import "time"

type Customer struct {
	CustomerID int64      `json:"customerid"`
	AuthID     int64      `json:"authid"`
	Username   *string    `json:"username,omitempty"`
	Fullname   *string    `json:"fullname,omitempty"`
	Email      string     `json:"email"`
	Address    *string    `json:"address,omitempty"`
	Phone      *string    `json:"phone,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}
