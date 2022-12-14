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
	MethodScamper
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
	ParentID  sql.NullInt32 `json:"-" db:"parent_id"`
	Method    int           `json:"method"`
	Type      int           `json:"type"`
	GroupID   sql.NullInt32 `json:"-" db:"group_id"`
	Options   []*Option     `json:"options"`
	ParentUid string        `json:"parent_uid" db:"parent_uid"`
	GroupUid  string        `json:"group_uid" db:"group_uid"`
	Username  string        `json:"user" db:"name"`
}

type PostResponse struct {
	Post
	ParentUid string `json:"parent_uid" db:"parent_uid"`
	GroupUid  string `json:"group_uid" db:"group_uid"`
	Username  string `json:"username" db:"name"`
}

// A post request model
//
// swagger:parameters getPost2
type PostBodyParams struct {
	// in: body
	Post *Post `json:"post"`
}

// swagger:parameters getPost removeOptions removePost
type GetPostParams struct {
	// required: true
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
	Limit    uint   `json:"limit"`
	GroupUid string `json:"group_uid"`
}

// swagger:parameters find
type FetchPostsWithTitle struct {
	// required: true
	// in: query
	FetchPostsParams

	// required: true
	// in: query
	Title string `json:"title"`
}

// swagger:parameters uploadPost
type PostFileBodyParams struct {
	// in: query
	GroupUid string `json:"group_uid"`

	// in: formData
	Title string `json:"title"`

	// in: formData
	ParentUID string `json:"parent_uid"`

	// required: true
	// in: formData
	// enum: 0,1,2,3
	Method int `json:"method"`

	// required: true
	// in: formData
	// enum: 0,1
	Type int `json:"type"`

	// required: true
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
