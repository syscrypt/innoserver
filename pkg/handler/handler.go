package handler

import (
	"context"
	"net/http"

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
	uniqueID
	SelectByUserID(ctx context.Context, id int) ([]*model.Post, error)
	SelectByParent(ctx context.Context, parent *model.Post) ([]*model.Post, error)
	GetByTitle(ctx context.Context, title string) (*model.Post, error)
	Persist(ctx context.Context, post *model.Post) error
	GetByUid(ctx context.Context, uid string) (*model.Post, error)
	SelectLatest(ctx context.Context, limit uint64) ([]*model.Post, error)
	SelectLatestOfGroup(ctx context.Context, group *model.Group, limit uint64) ([]*model.Post, error)
}

type groupRepository interface {
	uniqueID
	GetByUid(ctx context.Context, uid string) (*model.Group, error)
	Persist(ctx context.Context, group *model.Group) error
	AddUserToGroup(ctx context.Context, user *model.User, group *model.Group) error
	IsUserInGroup(ctx context.Context, user *model.User, group *model.Group) (bool, error)
	GetUsersInGroup(ctx context.Context, group *model.Group) ([]*model.User, error)
}

type uniqueID interface {
	UniqueIdExists(ctx context.Context, uid string) (bool, error)
}

type Handler struct {
	userRepo  userRepository
	postRepo  postRepository
	groupRepo groupRepository

	config *model.Config
	log    *logrus.Entry
	rlog   *logrus.Logger
}

func NewHandler(injections ...interface{}) *Handler {
	handler := &Handler{}
	for _, i := range injections {
		switch v := i.(type) {
		case userRepository:
			handler.userRepo = v
		case postRepository:
			handler.postRepo = v
		case groupRepository:
			handler.groupRepo = v
		case *model.Config:
			handler.config = v
		case [2]*logrus.Logger:
			handler.log = v[0].WithFields(logrus.Fields{})
			handler.rlog = v[1]
		}

	}
	return handler
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()
	s.log = s.log.WithFields(logrus.Fields{"url": r.URL.String()})
	r = r.WithContext(context.WithValue(r.Context(), "config", s.config))
	r = r.WithContext(context.WithValue(r.Context(), "user_repository", &s.userRepo))
	r = r.WithContext(context.WithValue(r.Context(), "group_repository", &s.groupRepo))
	r = r.WithContext(context.WithValue(r.Context(), "log", s.log))
	r = r.WithContext(context.WithValue(r.Context(), "rlog", s.rlog))
	swaggerRouter := router.PathPrefix("/swagger").Subrouter()
	swaggerRouter.Path("").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.Swagger))

	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.Path("/login").Methods("POST", "OPTIONS").HandlerFunc(errorWrapper(s.Login))
	authRouter.Path("/register").Methods("POST", "OPTIONS").HandlerFunc(errorWrapper(s.Register))

	postRouter := router.PathPrefix("/post").Subrouter()
	postRouter.Path("/upload").Methods("POST", "GET", "OPTIONS").HandlerFunc(errorWrapper(s.UploadPost))
	postRouter.Path("/get").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.GetPost))
	postRouter.Path("/getchildren").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.GetChildren))
	postRouter.Path("/selectlatest").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.FetchLatestPosts))
	postRouter.Use(authenticationMiddleware)

	groupRouter := router.PathPrefix("/group").Subrouter()
	groupRouter.Path("/create").Methods("POST", "OPTIONS").HandlerFunc(errorWrapper(s.CreateGroup))
	groupRouter.Path("/adduser").Methods("POST", "OPTIONS").HandlerFunc(errorWrapper(s.AddUserToGroup))
	groupRouter.Path("/listmembers").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.ListGroupMembers))
	groupRouter.Path("/info").Methods("POST", "OPTIONS").HandlerFunc(errorWrapper(s.GroupInfo))

	assetRouter := router.PathPrefix("/assets").Subrouter()
	assetRouter.PathPrefix("/images").Handler(http.StripPrefix("/assets/images",
		http.FileServer(http.Dir("assets/images/"))))
	assetRouter.PathPrefix("/videos").Handler(http.StripPrefix("/assets/videos",
		http.FileServer(http.Dir("assets/videos/"))))

	router.Use(corsMiddleware)
	router.Use(logMiddleware)
	authRouter.Use(keyMiddleware)
	postRouter.Use(keyMiddleware)
	groupRouter.Use(keyMiddleware)
	groupRouter.Use(authenticationMiddleware)
	postRouter.Use(authenticationMiddleware)
	groupRouter.Use(groupMiddleware)
	postRouter.Use(groupMiddleware)

	router.ServeHTTP(w, r)
}
