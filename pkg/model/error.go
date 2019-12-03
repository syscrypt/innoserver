package model

// An error response model
//
// swagger:response errorResponse
type ErrorResponse struct {
	Message string `json:"error_message"`
}
