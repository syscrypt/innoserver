package model

// User model
//
// swagger:model
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Imei     string `json:"imei"`
	Password string `json:"password"`
}

// An user request model
//
// swagger:parameters login
type LoginBodyParams struct {
	// The user to submit
	//
	// required: true
	// in: body
	User *User `json:"user"`
}
