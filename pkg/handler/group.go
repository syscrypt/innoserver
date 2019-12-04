package handler

import (
	"encoding/json"
	"errors"
	"net/http"

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
		return errors.New("missing parameter title in request query"), http.StatusBadRequest
	}
	group := &model.Group{}
	group.Title = title
	user, err := s.GetCurrentUser(r)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	group.AdminID = user.ID
	uid, err := generateUid(s.groupRepo, r)
	if err != nil || uid == "" {
		return errors.New("error while generating uid. " + err.Error()), http.StatusInternalServerError
	}
	group.UniqueID = uid
	err = s.groupRepo.Persist(r.Context(), group)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	err = s.groupRepo.AddUserToGroup(r.Context(), user, group)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	ret, err := json.Marshal(&model.UidResponse{UniqueID: uid})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	SetJsonHeader(w)
	w.Write(ret)
	return nil, http.StatusOK
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
	curUser, err := s.GetCurrentUser(r)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if curUser.Email == user.Email {
		return errors.New("cannot add requesting user to destination group"), http.StatusBadRequest
	}
	if curUser.ID != group.AdminID {
		return errors.New("requesting user is not the group admin"), http.StatusUnauthorized
	}
	inGroup, err := s.groupRepo.IsUserInGroup(r.Context(), user, group)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if inGroup {
		return errors.New("user is already member of group " + group.Title), http.StatusBadRequest
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
		return errors.New("parameter group_uid missing in query"), http.StatusBadRequest
	}
	group, err := s.groupRepo.GetByUid(r.Context(), group_uid)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	users, err := s.groupRepo.GetUsersInGroup(r.Context(), group)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	ret, err := json.Marshal(users)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	SetJsonHeader(w)
	w.Write(ret)
	return nil, http.StatusOK
}
