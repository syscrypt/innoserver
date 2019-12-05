package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"gitlab.com/innoserver/pkg/model"
)

func GetCurrentUser(r *http.Request, repo userRepository) (*model.User, error) {
	if username, ok := r.Context().Value("username").(string); ok {
		if username == "" {
			return nil, errors.New("no username provided")
		}
		return repo.GetByUsername(r.Context(), username)
	}
	return nil, errors.New("error fetching username in context values")
}

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

func generateUid(repo uniqueID, r *http.Request) (string, error) {
	for {
		uid, _ := uuid.NewRandom()
		exists, err := repo.UniqueIdExists(r.Context(), uid.String())
		if err != nil {
			return "", err
		}
		if !exists {
			return uid.String(), nil
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return "", errors.New("unknown error while generating uid")
}

func SetJsonHeader(w http.ResponseWriter) {
	w.Header().Set("content-type", "application/json")
}
