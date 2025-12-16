package model

import "time"

type Developer struct {
	DeveloperID   int64      `json:"developerid"`
	DeveloperName string     `json:"developername"`
	AuthID        *int64     `json:"authid,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}
