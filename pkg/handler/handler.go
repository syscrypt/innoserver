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
	SelectByParent(ctx context.Context, parent *model.Post) ([]*model.Post, error)
	SelectByUser(ctx context.Context, user *model.User) ([]*model.Post, error)
	GetByTitle(ctx context.Context, title string, limit int64) ([]*model.Post, error)
	Persist(ctx context.Context, post *model.Post) error
	GetByTitleInGroup(ctx context.Context, title string, group *model.Group, limit int64) ([]*model.Post, error)
	GetByUid(ctx context.Context, uid string) (*model.Post, error)
	SelectLatest(ctx context.Context, limit uint64) ([]*model.Post, error)
	SelectLatestOfGroup(ctx context.Context, group *model.Group, limit uint64) ([]*model.Post, error)
	AddOptions(ctx context.Context, post *model.Post, options []*model.Option) error
	RemoveOptions(ctx context.Context, post *model.Post) error
	SetOptions(ctx context.Context, post *model.Post, options []*model.Option) error
	SelectOptions(ctx context.Context, post *model.Post) ([]*model.Option, error)
	RemovePost(ctx context.Context, post *model.Post) error
}

type groupRepository interface {
	uniqueID
	GetByUid(ctx context.Context, uid string) (*model.Group, error)
	Persist(ctx context.Context, group *model.Group) error
	AddUserToGroup(ctx context.Context, user *model.User, group *model.Group) error
	IsUserInGroup(ctx context.Context, user *model.User, group *model.Group) (bool, error)
	GetUsersInGroup(ctx context.Context, group *model.Group) ([]*model.User, error)
	UpdateVisibility(ctx context.Context, group *model.Group) error
	SelectByUser(ctx context.Context, user *model.User) ([]*model.Group, error)
	RemoveGroup(ctx context.Context, group *model.Group) error
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

	router *mux.Router
}

func NewHandler(injections ...interface{}) *Handler {
	handler := &Handler{}
	handler.SetupRouter()

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

func (s *Handler) SetupRouter() {
	s.router = mux.NewRouter()

	s.router.Path("/config").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.GetConfig))

	userRouter := s.router.PathPrefix("/user").Subrouter()
	userRouter.Path("/info").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.UserInfo))

	swaggerRouter := s.router.PathPrefix("/swagger").Subrouter()
	swaggerRouter.Path("").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.Swagger))

	authRouter := s.router.PathPrefix("/auth").Subrouter()
	authRouter.Path("/login").Methods("POST", "OPTIONS").HandlerFunc(errorWrapper(s.Login))
	authRouter.Path("/register").Methods("POST", "OPTIONS").HandlerFunc(errorWrapper(s.Register))

	postRouter := s.router.PathPrefix("/post").Subrouter()
	postRouter.Path("/remove").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.RemovePost))
	postRouter.Path("/upload").Methods("POST", "GET", "OPTIONS").HandlerFunc(errorWrapper(s.UploadPost))
	postRouter.Path("/get").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.GetPost))
	postRouter.Path("/find").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.Find))
	postRouter.Path("/getchildren").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.GetChildren))
	postRouter.Path("/selectlatest").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.FetchLatestPosts))
	postRouter.Path("/setoptions").Methods("POST", "OPTIONS").HandlerFunc(errorWrapper(s.SetOptions))
	postRouter.Path("/addoptions").Methods("POST", "OPTIONS").HandlerFunc(errorWrapper(s.AddOptions))
	postRouter.Path("/removeoptions").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.RemoveOptions))
	postRouter.Use(authenticationMiddleware)

	groupRouter := s.router.PathPrefix("/group").Subrouter()
	groupRouter.Path("/join").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.JoinGroup))
	groupRouter.Path("/info").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.GroupInfo))

	inGroupRouter := groupRouter.PathPrefix("").Subrouter()
	inGroupRouter.Path("/create").Methods("POST", "OPTIONS").HandlerFunc(errorWrapper(s.CreateGroup))
	inGroupRouter.Path("/listmembers").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.ListGroupMembers))

	adminRouter := inGroupRouter.PathPrefix("").Subrouter()
	adminRouter.Path("/adduser").Methods("POST", "OPTIONS").HandlerFunc(errorWrapper(s.AddUserToGroup))
	adminRouter.Path("/setvisibility").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.SetVisibility))
	adminRouter.Path("/remove").Methods("GET", "OPTIONS").HandlerFunc(errorWrapper(s.RemoveGroup))

	assetRouter := s.router.PathPrefix("/assets").Subrouter()
	assetRouter.PathPrefix("/images").Handler(http.StripPrefix("/assets/images",
		http.FileServer(http.Dir("assets/images/"))))
	assetRouter.PathPrefix("/videos").Handler(http.StripPrefix("/assets/videos",
		http.FileServer(http.Dir("assets/videos/"))))

	s.router.Use(corsMiddleware)
	s.router.Use(logMiddleware)
	authRouter.Use(keyMiddleware)
	postRouter.Use(keyMiddleware)
	groupRouter.Use(keyMiddleware)
	userRouter.Use(keyMiddleware)
	userRouter.Use(authenticationMiddleware)
	groupRouter.Use(authenticationMiddleware)
	postRouter.Use(authenticationMiddleware)
	inGroupRouter.Use(groupMiddleware)
	postRouter.Use(groupMiddleware)
	adminRouter.Use(adminMiddleware)
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.log = s.log.WithFields(logrus.Fields{"url": r.URL.String()})
	r = r.WithContext(context.WithValue(r.Context(), "config", s.config))
	r = r.WithContext(context.WithValue(r.Context(), "user_repository", &s.userRepo))
	r = r.WithContext(context.WithValue(r.Context(), "group_repository", &s.groupRepo))
	r = r.WithContext(context.WithValue(r.Context(), "log", s.log))
	r = r.WithContext(context.WithValue(r.Context(), "rlog", s.rlog))
	s.router.ServeHTTP(w, r)
}
