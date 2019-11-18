package handler

import (
	"net/http"

	"github.com/sirupsen/logrus"

	_ "gitlab.com/innoserver/pkg/model"
)

// UploadPost swagger:route POST /uploadpost post uploadPost
//
// description: Takes, processes and persist posts data
//
// responses:
//     200: description: post was uploaded successfully
func (s *Handler) UploadPost(w http.ResponseWriter, r *http.Request) {
	logrus.Infoln("upload request initiated")
}
