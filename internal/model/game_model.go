package model

import "time"

type Game struct {
	GameID      int64      `json:"gameid"`
	DeveloperID int64      `json:"developerid"`
	Title       string     `json:"title"`
	Price       float64    `json:"price"`
	ReleaseDate *time.Time `json:"releasedate,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
