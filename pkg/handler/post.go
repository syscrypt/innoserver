package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
)

// UploadPost swagger:route POST /uploadpost post uploadPost
//
// description: Takes, processes and persist posts data
//
// responses:
//     200: description: post was uploaded successfully
//     400: description: bad request
func (s *Handler) UploadPost(w http.ResponseWriter, r *http.Request) {
	var post model.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		logrus.Errorln("uploadpost:" + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logrus.Print(post)
}

// UploadPost swagger:route POST /uploadpostfile uploadPostFile
//
// description: Takes, processes and persist posts data
//
// responses:
//     200: description: post was uploaded successfully
//     400: description: bad request
//     500: description: internal server error
func (s *Handler) UploadPostFile(w http.ResponseWriter, r *http.Request) {
	var maxSize int64
	fileType := r.PostFormValue("type")
	if fileType == "image" {
		maxSize = 10 << 20
	} else if fileType == "video" {
		maxSize = 10 << 40
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.UploadFile(r, maxSize, "post", fileType); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
