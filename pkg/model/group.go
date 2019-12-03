package model

// Group model
//
// swagger:model
type Group struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	AdminID  int    `json:"admin_id"`
	UniqueID string `json:"unique_id"`
}
