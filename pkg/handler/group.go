package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"gitlab.com/innoserver/pkg/model"
)

// CreateGroup swagger:route POST /group/create group createGroup
//
// Creates a new Group with the requester as admin
//
// responses:
//     200: uidResponse
//     400: description: bad request
//     500: description: server internal error
func (s *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) (error, int) {
	title := r.URL.Query().Get("title")
	if title == "" {
		return ErrMissingParam(w, "title", s.rlog)
	}
	s.log.WithField("group", title).Infoln("trying to create new group...")
	group := &model.Group{}
	group.Title = title
	user, err := GetCurrentUser(r)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	group.AdminID = user.ID
	uid, err := generateUid(s.groupRepo, r)
	if err != nil || uid == "" {
		return err, http.StatusInternalServerError
	}
	group.UniqueID = uid
	err = s.groupRepo.Persist(r.Context(), group)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	s.log.WithFields(logrus.Fields{
		"title": group.Title, "uid": group.UniqueID,
	}).Infoln("group created")
	newGroup, err := s.groupRepo.GetByUid(r.Context(), group.UniqueID)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	err = s.groupRepo.AddUserToGroup(r.Context(), user, newGroup)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	s.log.WithFields(logrus.Fields{
		"user": user.Name, "group": group.Title,
	}).Infoln("user added to group")
	return WriteJsonResp(w, &model.UidResponse{UniqueID: uid})
}

// AddUserToGroup swagger:route POST /group/adduser group addUserToGroup
//
// Adds a user (if exists) to a group (if exists)
//
// responses:
//     200: description: user successfully added to group
//     400: description: bad request
//     500: description: server internal error
func (s *Handler) AddUserToGroup(w http.ResponseWriter, r *http.Request) (error, int) {
	body := &model.UserGroupRelation{}
	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		return err, http.StatusBadRequest
	}
	user, err := s.userRepo.GetByEmail(r.Context(), body.Email)
	if err != nil {
		return err, http.StatusBadRequest
	}
	group, err := s.groupRepo.GetByUid(r.Context(), body.GroupUid)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	curUser, err := GetCurrentUser(r)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if curUser.Email == user.Email {
		return logResponse(w, "cannot add requesting user to group",
			s.rlog.WithFields(logrus.Fields{
				"user":        curUser.Name,
				"group_title": group.Title,
				"group_uid":   group.UniqueID,
			}), http.StatusUnauthorized)
	}
	if curUser.ID != group.AdminID {
		return logResponse(w, "operation not permitted, requesting user is not the group admin",
			s.rlog.WithFields(logrus.Fields{
				"user":        curUser.Name,
				"group_title": group.Title,
				"group_uid":   group.UniqueID,
			}), http.StatusUnauthorized)
	}
	inGroup, err := s.groupRepo.IsUserInGroup(r.Context(), user, group)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if inGroup {
		return logResponse(w, "user is already in group",
			s.rlog.WithFields(logrus.Fields{
				"user":        user.Name,
				"group_title": group.Title,
				"group_uid":   group.UniqueID,
			}), http.StatusBadRequest)
	}
	err = s.groupRepo.AddUserToGroup(r.Context(), user, group)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

// ListGroupMembers swagger:route GET /group/listmembers group listGroupMembers
//
// Returns a list with all members of specific group
//
// responses:
//     200: []User
//     400: description: bad request
//     500: description: server internal error
func (s *Handler) ListGroupMembers(w http.ResponseWriter, r *http.Request) (error, int) {
	group_uid := r.URL.Query().Get("group_uid")
	if group_uid == "" {
		return ErrMissingParam(w, "group_uid", s.rlog)
	}
	group, err := s.groupRepo.GetByUid(r.Context(), group_uid)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	users, err := s.groupRepo.GetUsersInGroup(r.Context(), group)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return WriteJsonResp(w, users)
}

// GroupInfo swagger:route POST /group/info group groupInfo
//
// Returns infos about specific group
//
// responses:
//     200: Group
//     400: description: bad request
//     500: description: server internal error
func (s *Handler) GroupInfo(w http.ResponseWriter, r *http.Request) (error, int) {
	groupUidReq := &model.GroupUniqueIdPostReq{}
	err := json.NewDecoder(r.Body).Decode(&groupUidReq.GroupUid)
	if err != nil {
		return err, http.StatusBadRequest
	}
	group, err := s.groupRepo.GetByUid(r.Context(), groupUidReq.GroupUid.Uid)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return WriteJsonResp(w, group)
}
