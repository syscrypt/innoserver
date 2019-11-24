package handler

import (
	"context"
	"errors"
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
	Persist(ctx context.Context, user *model.User) error
}

type Handler struct {
	userRepo userRepository

	swaggerSpecs []byte
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

	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.Path("/login").Methods("POST", "OPTIONS").HandlerFunc(s.Login)
	authRouter.Path("/register").Methods("POST", "OPTIONS").HandlerFunc(s.Register)

	postRouter := router.PathPrefix("/post").Subrouter()
	postRouter.Path("/uploadpost").Methods("POST", "OPTIONS").HandlerFunc(s.UploadPost)
	postRouter.Path("/uploadpostfile").Methods("POST", "OPTIONS").HandlerFunc(s.UploadPostFile)
	postRouter.Use(authenticationMiddleware)

	router.Use(corsMiddleware)
	router.ServeHTTP(w, r)
}

func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func authenticationMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("X-Auth-Token")
		if tokenStr != "" {
			claims := &model.Claims{}
			_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					errStr := "unexpected signing method"
					logrus.Errorln(errStr)
					return nil, errors.New(errStr)
				}
				return jwtKey, nil
			})
			if err != nil {
				logrus.Errorln("parsing incoming jw-token failed:", err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), "username", claims.Username))
			logrus.Println("user " + claims.Username + " authenticated")
			h.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	})
}
