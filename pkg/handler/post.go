package handler

import (
	"database/sql"
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
	user, err := GetCurrentUser(r)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	post := &model.Post{}
	err = initNewPostUpload(post, s, r)
	if err != nil {
		s.rlog.WithError(err)
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
