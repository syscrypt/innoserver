package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
	user, err := GetCurrentUser(r)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	post := &model.Post{}
	post.Title = r.FormValue("title")
	parentUid := r.FormValue("parent_uid")
	post.Method, err = strconv.Atoi(r.FormValue("method"))
	post.Type, err = strconv.Atoi(r.FormValue("type"))
	gUid := r.URL.Query().Get("group_uid")
	if gUid != "" {
		if group, err := s.groupRepo.GetByUid(r.Context(), gUid); err == nil {
			post.GroupID.Int32 = int32(group.ID)
			post.GroupID.Valid = true
		}
	}
	parent, err := s.postRepo.GetByUid(r.Context(), parentUid)
	if err != nil && err != sql.ErrNoRows {
		return err, http.StatusBadRequest
	}
	if parent.ID != 0 {
		post.ParentID.Int32 = int32(parent.ID)
		post.ParentID.Valid = true
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

	uid, err := generateUid(s.postRepo, r)
	if err != nil || uid == "" {
		return errors.New("error generating uid. " + err.Error()), http.StatusInternalServerError
	}
	post.UniqueID = uid
	if err := s.postRepo.Persist(r.Context(), post); err != nil {
		return err, http.StatusInternalServerError
	}
	ret, err := json.Marshal(&model.UidResponse{UniqueID: post.UniqueID})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	SetJsonHeader(w)
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
	parentPost, err := s.postRepo.GetByUid(r.Context(), parent)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	posts, err := s.postRepo.SelectByParent(r.Context(), parentPost)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	ret, err := json.Marshal(posts)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	SetJsonHeader(w)
	w.Write(ret)
	return nil, http.StatusOK
}

// FetchLatestPosts swagger:route GET /post/selectlatest post fetchLatestPosts
//
// Fetch all subposts of a specific parent post
// responses:
//    200: description: successfully returned a list of first X posts
//    400: description: Query error
//    500: description: Internal error
func (s *Handler) FetchLatestPosts(w http.ResponseWriter, r *http.Request) (error, int) {
	count := r.URL.Query().Get("limit")
	group_uid := r.URL.Query().Get("group_uid")
	icount, err := strconv.Atoi(count)
	var group *model.Group
	if count == "" || err != nil {
		return errors.New("parameter count missing in request query or wrong type"), http.StatusBadRequest
	}
	group, err = s.groupRepo.GetByUid(r.Context(), group_uid)
	if err != nil && err != sql.ErrNoRows {
		return err, http.StatusBadRequest
	}
	var posts []*model.Post
	if group.ID == 0 {
		posts, err = s.postRepo.SelectLatest(r.Context(), uint64(icount))
	} else {
		posts, err = s.postRepo.SelectLatestOfGroup(r.Context(), group, uint64(icount))
	}
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if err != nil {
		return err, http.StatusInternalServerError
	}
	postsStr, err := json.Marshal(posts)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	SetJsonHeader(w)
	w.Write(postsStr)
	return nil, http.StatusOK
}
