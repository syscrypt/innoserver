package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
)

// UploadPost swagger:route POST /post/upload post uploadPost
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
//     200: uidResponse
//     400: description: bad request
//     500: description: internal server error
func (s *Handler) UploadPost(w http.ResponseWriter, r *http.Request) (error, int) {
	var maxSize int64
	var path string
	var err error

	user, err := s.GetCurrentUser(r)

	post := &model.Post{}
	post.Title = r.FormValue("title")
	post.ParentUID = r.FormValue("parent_uid")
	post.Method, err = strconv.Atoi(r.FormValue("method"))
	post.Type, err = strconv.Atoi(r.FormValue("type"))
	if err != nil {
		return err, http.StatusInternalServerError
	}

	if post.Type < model.PostTypeImage || post.Type > model.PostTypeVideo {
		return errors.New("wrong type value for posted file"), http.StatusBadRequest
	}

	if post.Type == model.PostTypeImage {
		maxSize = s.config.MaxImageSize
	} else if post.Type == model.PostTypeVideo {
		maxSize = s.config.MaxVideoSize
	} else {
		return errors.New("wrong type value for posted file"), http.StatusBadRequest
	}

	if path, err = s.UploadFile(r, maxSize, "file", post.Type); err != nil {
		return err, http.StatusInternalServerError
	}
	post.Path = path
	post.UserID = user.ID

	for {
		uid, _ := uuid.NewRandom()
		logrus.Println(uid.String())
		var exists bool
		var err error
		if exists, err = s.postRepo.UniqueIdExists(r.Context(), uid.String()); err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			post.UniqueID = uid.String()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err := s.postRepo.Persist(r.Context(), post); err != nil {
		return err, http.StatusInternalServerError
	}

	uidResponse := &model.GetPostParams{}
	uidResponse.UniqueID = post.UniqueID
	ret, err := json.Marshal(uidResponse)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	w.Header().Set("content-type", "application/json")
	w.Write(ret)
	return nil, http.StatusOK
}

// GetPost swagger:route GET /post/get post getPost
//
// Fetch post over unique id
//
// responses:
//     200: description: postBody
//     400: description: bad request
//     500: description: server internal error
func (s *Handler) GetPost(w http.ResponseWriter, r *http.Request) (error, int) {
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		return errors.New("parameter uid missing"), http.StatusBadRequest
	}
	post, err := s.postRepo.GetByUid(r.Context(), uid)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	ret, err := json.Marshal(post)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	w.Header().Set("content-type", "application/json")
	w.Write(ret)
	return nil, http.StatusOK
}

// GetChildren swagger:route GET /post/getchildren post getChildren
//
// Fetch all subposts of a specific parent post
// responses:
//    200: description: successfully returned a list of subposts
func (s *Handler) GetChildren(w http.ResponseWriter, r *http.Request) (error, int) {
	parent := r.URL.Query().Get("parent_uid")
	if parent == "" {
		return errors.New("parameter parent_uid missing in request"), http.StatusBadRequest
	}
	posts, err := s.postRepo.SelectByParentUid(r.Context(), parent)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	ret, err := json.Marshal(posts)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	w.Header().Set("content-type", "application/json")
	w.Write(ret)
	return nil, http.StatusOK
}
