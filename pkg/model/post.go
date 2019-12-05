package model

import (
	"database/sql"
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
	ID        int           `json:"-"`
	UniqueID  string        `json:"unique_id" db:"unique_id"`
	Title     string        `json:"title"`
	UserID    int           `json:"-" db:"user_id"`
	Path      string        `json:"path"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
	ParentUID string        `json:"parent_uid" db:"parent_uid"`
	Method    int           `json:"method"`
	Type      int           `json:"type"`
	GroupID   sql.NullInt32 `json:"group_id" db:"group_id"`
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

// swagger:parameters fetchLatestPosts
type FetchPostsParams struct {
	// required: true
	// in: query
	Limit uint `json:"limit"`

	GroupUid string `json:"group_uid"`
}

// swagger:parameters uploadPost
type PostFileBodyParams struct {
	// in: query
	GroupUid string `json:"group_uid"`
	// in: formData
	Title string `json:"title"`
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

// A response containing a unique id
//
// swagger:response uidResponse
type UidResponse struct {
	UniqueID string `json:"unique_id"`
}
