package handler

import (
	"io/ioutil"
	"net/http"

	"gitlab.com/innoserver/pkg/model"
)

// Swagger swagger:route GET /swagger swagger
//
// Returns the swagger specifications
//
// responses:
//     200: description: Swagger specifications
func (s *Handler) Swagger(w http.ResponseWriter, r *http.Request) {
	if config, ok := r.Context().Value("config").(*model.Config); ok {
		if swaggerSpecs, err := ioutil.ReadFile(config.Swaggerfile); err == nil {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write(swaggerSpecs)
			return
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
}
