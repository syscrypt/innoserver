package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
)

// Login swagger:route POST /auth/login user login
//
// Verifies user credentials and generates jw-token
//
// responses:
//     200: loginResponse
//     400: description: bad request
//     500: description: server internal error
func (s *Handler) Login(w http.ResponseWriter, r *http.Request) {
	logrus.Info("login attempt made")
	var creds model.User
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		logrus.Errorln("login: error decoding json body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	expirationTime := time.Now().Add(5 * time.Hour)

	if _, err = s.userRepo.GetByUsername(r.Context(), creds.Name); err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	claims := &model.Claims{
		Username: creds.Name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lResp := &model.LoginResponse{
		Token: tokenString,
	}
	if ret, err := json.Marshal(lResp); err == nil {
		w.Write(ret)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var creds model.User

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		logrus.Errorln("register: error decoding json body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = s.userRepo.Persist(r.Context(), &creds)
	if err != nil {
		logrus.Errorln("register: could not persist user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	expiration := time.Now().Add(5 * time.Hour)

	claims := &model.Claims{
		Username: creds.Name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiration.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write([]byte(tokenString)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
