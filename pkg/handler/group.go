package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"gitlab.com/innoserver/pkg/model"
)

// CreateGroup swagger:route POST /group/creategroup group createGroup
//
// Creates a new Group with the requester as admin
//
// responses:
//     200: uidResponse
//     400: description: bad request
//     500: description: server internal error
func (s *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) (error, int) {
	group := &model.Group{}
	err := json.NewDecoder(r.Body).Decode(group)
	if err != nil {
		return err, http.StatusBadRequest
	}
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
	ret, err := json.Marshal(&model.UidResponse{UniqueID: uid})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	w.Write(ret)
	return nil, http.StatusOK
}
