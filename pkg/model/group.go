package model

// Group model
//
// swagger:model
type Group struct {
	ID       int    `json:"-"`
	Title    string `json:"title"`
	AdminID  int    `json:"-" db:"admin_id"`
	UniqueID string `json:"-" db:"unique_id"`
}

// An user request model
//
// swagger:parameters createGroup
type CreatePostBodyParams struct {
	// The user to submit
	//
	// required: true
	// in: body
	Group *Group `json:"group"`
}
