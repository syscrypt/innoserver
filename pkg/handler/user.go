package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

func hashAndSalt(passwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(passwd, bcrypt.DefaultCost)
	if err != nil {
		logrus.Println(err)
	}
	return string(hash)
}

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

func (s *Handler) GetCurrentUser(r *http.Request) (*model.User, error) {
	if username, ok := r.Context().Value("username").(string); ok {
		if username == "" {
			return nil, errors.New("no username provided")
		}
		return s.userRepo.GetByUsername(r.Context(), username)
	}

	return nil, errors.New("error fetching username in context values")
}
