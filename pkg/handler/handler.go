package handler

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"gitlab.com/innoserver/pkg/model"
)

type userRepository interface {
	GetByUsername(ctx context.Context, name string) (*model.User, error)
	Persist(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type postRepository interface {
	SelectByUserID(ctx context.Context, id int) ([]*model.Post, error)
	SelectByParentUid(ctx context.Context, uid string) ([]*model.Post, error)
	GetByTitle(ctx context.Context, title string) (*model.Post, error)
	Persist(ctx context.Context, post *model.Post) error
	GetByUid(ctx context.Context, uid string) (*model.Post, error)
	UniqueIdExists(ctx context.Context, uid string) (bool, error)
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
	postRouter.Path("/upload").Methods("POST", "OPTIONS").HandlerFunc(s.UploadPost)
	postRouter.Path("/get").Methods("GET", "OPTIONS").HandlerFunc(s.GetPost)
	postRouter.Path("/getchildren").Methods("GET", "OPTIONS").HandlerFunc(s.GetChildren)
	postRouter.Use(authenticationMiddleware)

	assetRouter := router.PathPrefix("/assets").Subrouter()
	assetRouter.PathPrefix("/images").Handler(http.StripPrefix("/assets/images",
		http.FileServer(http.Dir("assets/images/"))))
	assetRouter.PathPrefix("/videos").Handler(http.StripPrefix("/assets/videos",
		http.FileServer(http.Dir("assets/videos/"))))

	router.Use(corsMiddleware)
	authRouter.Use(keyMiddleware)
	postRouter.Use(keyMiddleware)

	router.ServeHTTP(w, r)
}
