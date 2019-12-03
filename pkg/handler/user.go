package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
	"golang.org/x/crypto/bcrypt"
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
	logrus.Infoln(r.URL.String()+": login attempt by user", creds.Name)

	user, err := s.userRepo.GetByEmail(r.Context(), creds.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)) != nil {
		return err, http.StatusUnauthorized
	}

	creds = user
	token, err := s.generateToken(creds)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	ret, _ := json.Marshal(token)
	w.Write(ret)
	return nil, http.StatusOK
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
	logrus.Println(r.URL.String()+": Registration attempt made by new user", creds.Name)

	creds.Password = hashAndSalt([]byte(creds.Password))
	err = s.userRepo.Persist(r.Context(), creds)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	token, err := s.generateToken(creds)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	ret, _ := json.Marshal(token)
	w.Write(ret)
	return nil, http.StatusOK
}
