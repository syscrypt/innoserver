package model

// An user request model
//
// swagger:parameters login
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Imei     string `json:"imei"`
	Password string `json:"password"`
}
