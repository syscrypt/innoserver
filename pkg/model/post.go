package model

import (
	"os"
	"time"
)

const (
	PostTypeImage = iota
	PostTypeVideo
)

const (
	Method101 = iota
	MethodLotusBlossum
)

// A post request model
//
// swagger:model
type Post struct {
	ID        int       `json:"-"`
	UniqueID  string    `json:"unique_id" db:"unique_id"`
	Title     string    `json:"title"`
	UserID    int       `json:"-" db:"user_id"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"-" db:"created_at"`
	ParentUID string    `json:"parent_uid" db:"parent_uid"`
	Method    int       `json:"method"`
	Type      int       `json:"type"`
}

// A post request model
//
// swagger:parameters getPost2
type PostBodyParams struct {
	// in: body
	Post *Post `json:"post"`
}

// swagger:parameters getPost
type GetPostParams struct {
	// in: query
	UniqueID string `json:"uid"`
}

// swagger:parameters getChildren
type GetChildrenParams struct {
	// in: query
	ParentUid string `json:"parent_uid"`
}

// swagger:parameters uploadPost
type PostFileBodyParams struct {
	// in: formData
	Title string `json:"title"`
	// in: formData
	UniqueID string `json:"unique_id"`
	// in: formData
	ParentUID string `json:"parent_uid"`
	// in: formData
	// enum: 0,1
	Method int `json:"method"`
	// in: formData
	// enum: 0,1
	Type int `json:"type"`
	// in: formData
	// swagger:file
	// name: file
	File *os.File `json:"file"`
}
