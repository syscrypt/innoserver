package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
)

// UploadPost swagger:route POST /post/upload post uploadPost
//
//   <p>Takes, processes and persist posts data
//   A post file request model.
//   Parameter "Method" is an integer and takes following numbers:</p>
//     <ul><li>0: 101 Method</li>
//     <li>1: Lotus Blossum</li>
//     <li>2: Scamper</li>
//     <li>3: Diese eine andere, ka...</li></ul>
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
	user, err := GetCurrentUser(r)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	post := &model.Post{}
	err = initNewPostUpload(post, s, r)
	if err != nil {
		return logResponse(w, "upload failed", s.rlog.WithError(err), http.StatusBadRequest)
	}
	s.log.WithFields(logrus.Fields{
		"title": post.Title, "user": user.Name,
	}).Infoln("trying to upload new post...")

	if post.Type < model.PostTypeImage || post.Type > model.PostTypeVideo {
		return logResponse(w, "wrong type for post",
			s.rlog.WithFields(logrus.Fields{
				"type": post.Type,
			}), http.StatusBadRequest)
	}
	maxSize := determineMaxPostSize(post, s)
	path, _, err := s.UploadFile(r, maxSize, "file", post.Type)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	post.Path = path
	post.UserID = user.ID
	uid, err := generateUid(s.postRepo, r)
	if err != nil || uid == "" {
		return err, http.StatusInternalServerError
	}
	post.UniqueID = uid
	if err := s.postRepo.Persist(r.Context(), post); err != nil {
		return err, http.StatusInternalServerError
	}
	s.log.WithFields(logrus.Fields{
		"title": post.Title, "user": user.Name,
	}).Infoln("post uploaded successfully")
	return WriteJsonResp(w, &model.UidResponse{UniqueID: post.UniqueID})
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
		return ErrMissingParam(w, "uid", s.rlog)
	}
	post, err := s.postRepo.GetByUid(r.Context(), uid)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	s.log.WithFields(logrus.Fields{
		"title": post.Title, "uid": post.UniqueID,
	}).Infoln("fetching post")
	return WriteJsonResp(w, post)
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
	s.log.WithFields(logrus.Fields{
		"parent":       parent,
		"parent_title": parentPost.Title,
		"children":     len(posts),
	}).Infoln("fetching child posts")
	return WriteJsonResp(w, posts)
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
	s.log.WithFields(logrus.Fields{
		"limit":     count,
		"group_uid": group_uid,
	}).Infoln("fetching latest post")
	if count == "" || err != nil {
		return logResponse(w, "missing parameter in request query, or wrong type",
			s.rlog.WithField("limit", count), http.StatusBadRequest)
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
	return WriteJsonResp(w, posts)
}

// SetOptions swagger:route POST /post/setoptions post setOptions
//
// Set a List of options for a post
//
// responses:
//    200: description: successfully updated posts options
func (s *Handler) SetOptions(w http.ResponseWriter, r *http.Request) (error, int) {
	post_uid := r.URL.Query().Get("post_uid")
	var options []*model.Option
	err := json.NewDecoder(r.Body).Decode(&options)
	if err != nil {
		return err, http.StatusBadRequest
	}
	user, err := GetCurrentUser(r)
	if err != nil {
		return err, http.StatusUnauthorized
	}
	if post_uid == "" {
		return ErrMissingParam(w, "post_uid", s.rlog)
	}
	post, err := s.postRepo.GetByUid(r.Context(), post_uid)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if user.ID != post.UserID {
		return err, http.StatusUnauthorized
	}
	for _, v := range options {
		v.PostUid = post_uid
	}
	err = s.postRepo.SetOptions(r.Context(), post, options)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

// Find swagger:route GET /post/find post find
//
// Fetch latest posts with title
//
// responses:
//    200: []Post
//    400: description: Query error
//    500: description: Internal error
func (s *Handler) Find(w http.ResponseWriter, r *http.Request) (error, int) {
	title := r.URL.Query().Get("title")
	count := r.URL.Query().Get("limit")
	group_uid := r.URL.Query().Get("group_uid")
	icount, err := strconv.Atoi(count)
	var group *model.Group
	s.log.WithFields(logrus.Fields{
		"limit":     count,
		"group_uid": group_uid,
		"title":     title,
	}).Infoln("fetching posts with title")
	if count == "" || err != nil {
		return logResponse(w, "missing parameter in request query, or wrong type",
			s.rlog.WithField("limit", count), http.StatusBadRequest)
	}
	group, err = s.groupRepo.GetByUid(r.Context(), group_uid)
	if err != nil && err != sql.ErrNoRows {
		return err, http.StatusBadRequest
	}
	var posts []*model.Post
	if group.ID == 0 {
		posts, err = s.postRepo.GetByTitle(r.Context(), title, int64(icount))
	} else {
		posts, err = s.postRepo.GetByTitleInGroup(r.Context(), title, group, int64(icount))
	}
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return WriteJsonResp(w, posts)
}

// AddOptions swagger:route POST /post/addoptions post AddOptions
//
// Set a List of options for a post
//
// responses:
//    200: description: successfully added posts options
func (s *Handler) AddOptions(w http.ResponseWriter, r *http.Request) (error, int) {
	post_uid := r.URL.Query().Get("post_uid")
	var options []*model.Option
	err := json.NewDecoder(r.Body).Decode(&options)
	if err != nil {
		return err, http.StatusBadRequest
	}
	user, err := GetCurrentUser(r)
	if post_uid == "" {
		return ErrMissingParam(w, "post_uid", s.rlog)
	}
	post, err := s.postRepo.GetByUid(r.Context(), post_uid)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if user.ID != post.UserID {
		return err, http.StatusUnauthorized
	}
	for _, v := range options {
		v.PostUid = post_uid
	}
	err = s.postRepo.AddOptions(r.Context(), post, options)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

// RemoveOptions swagger:route GET /post/removeoptions post removeOptions
//
// Remove all options from a specific post
//
// responses:
//    200: description: successfully removed posts options
func (s *Handler) RemoveOptions(w http.ResponseWriter, r *http.Request) (error, int) {
	postUid := r.URL.Query().Get("uid")

	user, err := GetCurrentUser(r)
	if postUid == "" {
		return ErrMissingParam(w, "uid", s.rlog)
	}
	post, err := s.postRepo.GetByUid(r.Context(), postUid)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if user.ID != post.UserID {
		return err, http.StatusUnauthorized
	}
	return nil, http.StatusOK
}

// RemoveOptions swagger:route GET /post/remove post removePost
//
// Remove post
//
// responses:
//    200: description: successfully removed post
func (s *Handler) RemovePost(w http.ResponseWriter, r *http.Request) (error, int) {
	postUid := r.URL.Query().Get("uid")

	user, err := GetCurrentUser(r)
	if postUid == "" {
		return ErrMissingParam(w, "uid", s.rlog)
	}
	post, err := s.postRepo.GetByUid(r.Context(), postUid)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if user.ID != post.UserID {
		return err, http.StatusUnauthorized
	}
	err = s.postRepo.RemovePost(r.Context(), post)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}
