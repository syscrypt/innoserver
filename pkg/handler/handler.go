package handler

import (
	"context"
	"net/http"

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

	activeUsers map[string]*model.User
}

func NewHandler(injections ...interface{}) *Handler {
	handler := &Handler{}
	handler.activeUsers = make(map[string]*model.User)

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
	r = r.WithContext(context.WithValue(r.Context(), "activeUsers", s.activeUsers))

	router.Path("/login").Methods("GET").HandlerFunc(s.Login)
	router.Path("/uploadpost").Methods("POST").HandlerFunc(s.UploadPost)
	router.Path("/uploadpostfile").Methods("POST").HandlerFunc(s.UploadPostFile)

	router.Use(authenticationMiddleware)
	router.ServeHTTP(w, r)
}

func authenticationMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if activeUsers, ok := r.Context().Value("activeUsers").(map[string]*model.User); ok {
			jwt := r.Header.Get("jwt")
			if jwt != "" {
				user := activeUsers[jwt]
				if user != nil {
					r = r.WithContext(context.WithValue(r.Context(), "user", user))
					h.ServeHTTP(w, r)
					return
				}
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
		h.ServeHTTP(w, r)
	})
}
