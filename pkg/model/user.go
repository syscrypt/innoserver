package model

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IMEI     string `json:"imei"`
	Password string `json:"password"`
}
