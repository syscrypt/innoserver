package handler

import (
	"context"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
)

// TODO load from config
var jwtKey = []byte("secret")

type userRepository interface {
	GetByUsername(ctx context.Context, name string) (*model.User, error)
}

type Handler struct {
	userRepo userRepository
}

func NewHandler(injections ...interface{}) *Handler {
	handler := &Handler{}

	for _, i := range injections {
		switch v := i.(type) {
		case userRepository:
			logrus.Println("injectded user repository")
			handler.userRepo = v
		}
	}

	return handler
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()

	router.Path("/login").Methods("GET", "POST").HandlerFunc(s.login)

	router.ServeHTTP(w, r)
}
