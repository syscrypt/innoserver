package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
)

func (s *Handler) generateToken(user *model.User) (*model.TokenResponse, error) {
	response := &model.TokenResponse{}
	var err error
	expirationTime := time.Now().Add(5 * time.Hour)

	claims := &model.Claims{
		Username: user.Name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	response.Token, err = token.SignedString([]byte(s.config.JwtSecret))
	if err != nil {
		return nil, err
	}
	response.Name = user.Name

	return response, nil
}

// Login swagger:route POST /auth/login user login
//
// Verifies user credentials and generates jw-token
//
// responses:
//     200: tokenResponse
//     400: description: bad request
//     401: description: wrong user credentials
//     500: description: server internal error
func (s *Handler) Login(w http.ResponseWriter, r *http.Request) {
	logrus.Info("login attempt made")
	creds := &model.User{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		logrus.Errorln("login: error decoding json body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user, err := s.userRepo.GetByEmail(r.Context(), creds.Email); err != nil ||
		user.Password != creds.Password {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else {
		creds = user
	}

	if token, err := s.generateToken(creds); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		ret, _ := json.Marshal(token)
		w.Write(ret)
		return
	} else {
		logrus.Errorln(err.Error())
	}

	w.WriteHeader(http.StatusInternalServerError)
}

// Register swagger:route POST /auth/register user register
//
// Persists a user in the database and generates jw-token
//
// responses:
//     200: tokenResponse
//     400: description: bad request
//     500: description: server internal error
func (s *Handler) Register(w http.ResponseWriter, r *http.Request) {
	logrus.Info("registration attempt made")
	creds := &model.User{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		logrus.Error("register: error decoding json body", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.userRepo.Persist(r.Context(), creds)
	if err != nil {
		logrus.Errorln("register: could not persist user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if token, err := s.generateToken(creds); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		ret, _ := json.Marshal(token)
		w.Write(ret)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}

func (s *Handler) GetCurrentUser(r *http.Request) (*model.User, error) {
	if username, ok := r.Context().Value("username").(string); ok {
		if username == "" {
			return nil, errors.New("no username provided")
		}
		return s.userRepo.GetByUsername(r.Context(), username)
	}

	return nil, errors.New("error fetching username in context values")
}
