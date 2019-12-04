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

// swagger:model
type UserGroupRelation struct {
	Email    string `json:"email"`
	GroupUid string `json:"group_uid"`
}

// swagger:parameters addUserToGroup
type AddUserToGroupRequestBody struct {
	// The user to be added
	// required: true
	// in: body
	Relation *UserGroupRelation `json:"relation"`
}

// swagger:parameters listGroupMembers
type ListMembersParams struct {
	// required: true
	// in: query
	GroupUid string `json:"group_uid"`
}

// swagger:parameters createGroup
type CreateGroupRequestBody struct {
	// in: query
	Title string `json:"title"`
}
