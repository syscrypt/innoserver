package model

import (
	"time"
)

// A post request model
//
// swagger:parameters post
type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	UserID    int       `json:"userID" db:"user_id"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
