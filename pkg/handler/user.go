package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
)

// A response for the login routine
//
// swagger:response loginResponse
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Login swagger:route POST /register user register
//
// Verifies user credentials and generates jw-token
//
// responses:
//     200: loginResponse
//     400: description: bad request
//     500: description: server internal error
func (s *Handler) Register(w http.ResponseWriter, r *http.Request) {
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

	claims := &Claims{
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

	if _, err := w.Write([]byte(tokenString)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
