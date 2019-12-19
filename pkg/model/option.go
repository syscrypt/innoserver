package model

// swagger:model
type Option struct {
	Key     string `json:"key" db:"opt_key"`
	Value   string `json:"value" db:"opt_value"`
	PostUid string `json:"-" db:"post_uid"`
}

// swagger:parameters setOptions AddOptions
type AddOptionReqBody struct {
	// in: query
	// required: true
	PostUid string `json:"post_uid"`

	// in: body
	// required: true
	Options []*Option `json:"options"`
}
