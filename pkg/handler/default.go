package handler

import (
	"net/http"

	"gitlab.com/innoserver/pkg/model"
)

// Config swagger:route GET /config getConfig
//
// Returns relevant server settings
//
// responses:
//     200: Configuration
func (s *Handler) GetConfig(w http.ResponseWriter, r *http.Request) (error, int) {
	config := &model.Configuration{
		MaxImageSize: s.config.MaxImageSize,
		MaxVideoSize: s.config.MaxVideoSize,
		ImagePath:    s.config.ImagePath,
		VideoPath:    s.config.VideoPath,
	}
	return WriteJsonResp(w, config)
}
