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
func (s *Handler) Swagger(w http.ResponseWriter, r *http.Request) (error, int) {
	var err error
	if config, ok := r.Context().Value("config").(*model.Config); ok {
		if swaggerSpecs, err := ioutil.ReadFile(config.Swaggerfile); err == nil {
			w.Header().Set("content-type", "application/json")
			w.Write(swaggerSpecs)
			return nil, http.StatusOK
		}
	}
	return err, http.StatusInternalServerError
}
