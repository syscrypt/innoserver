package handler

import (
	"net/http"
)

// Response for login routine
//
// swagger:response swaggerResponse
type swaggerResponse struct {
	Token string
}

// Swagger swagger:route GET /swagger swagger
//
// Returns the swagger specifications
//
// responses:
//     200: swaggerResponse
func (s *Handler) Swagger(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write(s.swaggerSpecs)
}
