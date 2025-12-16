package model

import "time"

type Genre struct {
	GenreID   int64      `json:"genreid"`
	GenreName string     `json:"genrename"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
