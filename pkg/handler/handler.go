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

type userRepository interface {
	GetByUsername(ctx context.Context, name string) (*model.User, error)
	Persist(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type postRepository interface {
	SelectByUserID(ctx context.Context, id int) ([]*model.Post, error)
	GetByTitle(ctx context.Context, title string) (*model.Post, error)
	Persist(ctx context.Context, post *model.Post) error
}

type Handler struct {
	userRepo userRepository
	postRepo postRepository

	config *model.Config
}

func NewHandler(injections ...interface{}) *Handler {
	handler := &Handler{}

	for _, i := range injections {
		switch v := i.(type) {
		case userRepository:
			handler.userRepo = v
		case postRepository:
			handler.postRepo = v
		case *model.Config:
			handler.config = v
		}
	}

	return handler
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()
	r = r.WithContext(context.WithValue(r.Context(), "config", s.config))
	swaggerRouter := router.PathPrefix("/swagger").Subrouter()
	swaggerRouter.Path("").Methods("GET", "OPTIONS").HandlerFunc(s.Swagger)

	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.Path("/login").Methods("POST", "OPTIONS").HandlerFunc(s.Login)
	authRouter.Path("/register").Methods("POST", "OPTIONS").HandlerFunc(s.Register)

	postRouter := router.PathPrefix("/post").Subrouter()
	postRouter.Path("/uploadpost").Methods("POST", "OPTIONS").HandlerFunc(s.UploadPost)
	postRouter.Use(authenticationMiddleware)

	router.Use(corsMiddleware)
	authRouter.Use(keyMiddleware)
	postRouter.Use(keyMiddleware)

	router.ServeHTTP(w, r)
}

func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config, _ := r.Context().Value("config").(*model.Config)
		w.Header().Set("Access-Control-Allow-Origin", config.AccessControlAllowOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", config.AccessControlAllowCredentials)
		w.Header().Set("Access-Control-Allow-Methods", config.AccessControlAllowMethods)
		w.Header().Set("Access-Control-Allow-Headers", config.AccessControlAllowHeaders)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func keyMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config, _ := r.Context().Value("config").(*model.Config)
		if r.Header.Get("API_KEY") != config.ApiKey && config.ApiKey != "" {
			w.WriteHeader(http.StatusUnauthorized)
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
				if config, ok := r.Context().Value("config").(*model.Config); ok {
					return []byte(config.JwtSecret), nil
				}
				return nil, nil
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
