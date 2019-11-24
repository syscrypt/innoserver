package model

import (
	"time"
)

// A post request model
//
// swagger:model
type Post struct {
	ID        int       `json:"-"`
	Title     string    `json:"title"`
	UserID    int       `json:"-" db:"user_id"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_id" db:"created_at"`
}

// A post request model
//
// swagger:parameters uploadPost
type PostBodyParams struct {
	// The post to submit
	//
	// required: true
	// in: body
	Post *Post `json:"post"`
}
