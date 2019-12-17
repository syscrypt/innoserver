package model

// swagger:model
type Option struct {
	ID      int    `json:"-"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	PostUid string `json:"post_uid" db:"post_uid"`
}

// swagger:parameters setOptions
type AddOptionReqBody struct {
	// in: body
	// required: true
	Options []*Option `json:"options"`
}
