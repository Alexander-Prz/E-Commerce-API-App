package model

import "time"

type Auth struct {
	AuthID       int64      `json:"authid"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // never JSON-encode
	Role         string     `json:"role"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`

	EmailVerified bool `json:"email_verified"`
}
