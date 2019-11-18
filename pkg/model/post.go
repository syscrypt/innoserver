package model

// A post request model
//
// swagger:parameters post
type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}
