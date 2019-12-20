package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"gitlab.com/innoserver/pkg/model"
)

// Login swagger:route POST /auth/login user login
//
// Verifies user credentials and generates jw-token
//
// responses:
//     200: tokenResponse
//     400: description: bad request
//     401: description: wrong user credentials
//     500: description: server internal error
func (s *Handler) Login(w http.ResponseWriter, r *http.Request) (error, int) {
	creds := &model.User{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		return err, http.StatusBadRequest
	}
	s.log.WithField("user", creds.Email).Infoln("login attempt made")
	user, err := s.userRepo.GetByEmail(r.Context(), creds.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)) != nil {
		return logResponse(w, "password validation failed",
			s.rlog.WithFields(logrus.Fields{
				"name":  creds.Name,
				"email": creds.Email,
			}),
			http.StatusUnauthorized)
	}
	creds = user
	return WriteTokenResp(w, creds, []byte(s.config.JwtSecret))
}

// Register swagger:route POST /auth/register user register
//
// Persists a user in the database and generates jw-token
//
// responses:
//     200: tokenResponse
//     400: description: bad request
//     500: description: server internal error
func (s *Handler) Register(w http.ResponseWriter, r *http.Request) (error, int) {
	creds := &model.User{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		return err, http.StatusBadRequest
	}
	s.log.WithField("user", creds.Name).Infoln("registration attempt made")
	creds.Password = hashAndSalt([]byte(creds.Password))
	err = s.userRepo.Persist(r.Context(), creds)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return WriteTokenResp(w, creds, []byte(s.config.JwtSecret))
}

// Register swagger:route GET /user/info user userInfo
//
// Fetch user Information with groups and posts
//
// responses:
//     200: UserWithPostsGroups
//     500: description: server internal error
func (s *Handler) UserInfo(w http.ResponseWriter, r *http.Request) (error, int) {
	user, err := GetCurrentUser(r)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	groups, err := s.groupRepo.SelectByUser(r.Context(), user)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	posts, err := s.postRepo.SelectByUser(r.Context(), user)
	resp := &model.UserWithPostsGroups{
		User:   *user,
		Groups: groups,
		Posts:  posts,
	}
	resp.User.Password = ""
	return WriteJsonResp(w, resp)
}
