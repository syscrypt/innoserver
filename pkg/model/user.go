package model

import (
	jwt "github.com/dgrijalva/jwt-go"
)

// User model
//
// swagger:model
type User struct {
	ID       int    `json:"-" db:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Imei     string `json:"imei"`
	Password string `json:"password"`
}

// swagger:model
type UserWithPostsGroups struct {
	User
	Groups []*Group `json:"groups"`
	Posts  []*Post  `json:"posts"`
}

// An user request model
//
// swagger:parameters login register
type LoginBodyParams struct {
	// The user to submit
	//
	// required: true
	// in: body
	User *User `json:"user"`
}

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

// Response for login and register routine
//
// swagger:response tokenResponse
type TokenResponse struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}
