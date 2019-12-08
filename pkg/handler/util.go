package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"gitlab.com/innoserver/pkg/model"
)

func GetCurrentUser(r *http.Request) (*model.User, error) {
	if user, ok := r.Context().Value("user").(*model.User); ok {
		return user, nil
	}
	return nil, errors.New("error fetching user in context value")
}

func hashAndSalt(passwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(passwd, bcrypt.DefaultCost)
	if err != nil {
		logrus.Println(err)
	}
	return string(hash)
}

func GenerateToken(user *model.User, secret []byte) (*model.TokenResponse, error) {
	response := &model.TokenResponse{}
	var err error
	expirationTime := time.Now().Add(5 * time.Hour)
	claims := &model.Claims{
		Email: user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	response.Token, err = token.SignedString(secret)
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
			return "", errors.New("error while generating uid" + err.Error())
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

func logResponse(w http.ResponseWriter, msg string, entry *logrus.Entry, status int) (error, int) {
	SetJsonHeader(w)
	w.WriteHeader(status)
	entry.Error(msg)
	return nil, status
}

func ErrMissingParam(w http.ResponseWriter, param string, log *logrus.Logger) (error, int) {
	SetJsonHeader(w)
	w.WriteHeader(http.StatusBadRequest)
	log.WithFields(logrus.Fields{
		"param": param,
	}).Error("missing parameter in request")
	return nil, http.StatusBadRequest
}

func WriteJsonResp(w http.ResponseWriter, msg interface{}) (error, int) {
	ret, err := json.Marshal(msg)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	SetJsonHeader(w)
	w.Write(ret)
	return nil, http.StatusOK
}

func WriteTokenResp(w http.ResponseWriter, user *model.User, secret []byte) (error, int) {
	token, err := GenerateToken(user, secret)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return WriteJsonResp(w, token)
}
