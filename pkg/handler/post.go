package handler

import (
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
)

// UploadPost swagger:route POST /post/uploadpost uploadPost
//
//   <p>Takes, processes and persist posts data
//   A post file request model.
//   Parameter "Method" is an integer and takes following numbers:</p>
//     <ul><li>0: 101 Method</li>
//     <li>1: Lotus Blossum</li></ul>
//   <p>Type is an integer and describes the file type:</p>
//     <ul><li>0: image</li>
//     <li>1: video</li></ul>
//
//
// consumes:
//     multipart/form-data
//
// responses:
//     200: description: post was uploaded successfully
//     400: description: bad request
//     500: description: internal server error
func (s *Handler) UploadPost(w http.ResponseWriter, r *http.Request) {
	var maxSize int64
	var path string
	var err error

	user, err := s.GetCurrentUser(r)

	post := &model.Post{}
	post.UniqueID = r.FormValue("unique_id")
	post.Title = r.FormValue("title")
	post.ParentUID = r.FormValue("parent_uid")
	post.Method, err = strconv.Atoi(r.FormValue("method"))
	post.Type, err = strconv.Atoi(r.FormValue("type"))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Errorln(err.Error())
		return
	}

	if post.Type < model.PostTypeImage || post.Type > model.PostTypeVideo {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Errorln("uploadpost: wrong value for type")
		return
	}

	if post.Type == model.PostTypeImage {
		maxSize = 10 << 20
	} else if post.Type == model.PostTypeVideo {
		maxSize = 10 << 24
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if path, err = s.UploadFile(r, maxSize, "file", post.Type); err != nil {
		logrus.Errorln(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	post.Path = path
	post.UserID = user.ID

	if err := s.postRepo.Persist(r.Context(), post); err != nil {
		logrus.Errorln(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
