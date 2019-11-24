package model

import (
	jwt "github.com/dgrijalva/jwt-go"
)

// User model
//
// swagger:model
type User struct {
	ID       int    `json:"-"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Imei     string `json:"imei"`
	Password string `json:"password"`
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
	Username string `json:"username"`
	jwt.StandardClaims
}

// Response for login and register routine
//
// swagger:response tokenResponse
type TokenResponse struct {
	Token string `json:"token"`
}
