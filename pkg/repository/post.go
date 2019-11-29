package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"gitlab.com/innoserver/pkg/model"
)

const (
	dqlAllPostsByUserID = `SELECT * FROM posts WHERE user_id = ?`
	dqlGetPostByTitle   = `SELECT * FROM posts WHERE title = ?`
	dqlGetPostByUid     = `SELECT * FROM posts WHERE unique_id = ? LIMIT 1`
	persistPost         = `INSERT INTO posts
						   (title, user_id, path, parent_uid,
							method, type, unique_id)
						   VALUES (?, ?, ?, ?, ?, ?, ?)`
)

type postRepository struct {
	persist        *sqlx.Stmt
	selectByUserID *sqlx.Stmt
	getByTitle     *sqlx.Stmt
	getByUID       *sqlx.Stmt
}

func NewPostRepository(db *sqlx.DB) (*postRepository, error) {
	ctx := context.Background()

	ctxPersistPost, err := db.PreparexContext(ctx, persistPost)
	if err != nil {
		return nil, err
	}
	ctxSelectByUserID, err := db.PreparexContext(ctx, dqlAllPostsByUserID)
	if err != nil {
		return nil, err
	}
	ctxGetByTitle, err := db.PreparexContext(ctx, dqlGetPostByTitle)
	if err != nil {
		return nil, err
	}
	ctxGetByUID, err := db.PreparexContext(ctx, dqlGetPostByUid)
	if err != nil {
		return nil, err
	}

	return &postRepository{
		persist:        ctxPersistPost,
		selectByUserID: ctxSelectByUserID,
		getByTitle:     ctxGetByTitle,
		getByUID:       ctxGetByUID,
	}, err
}

func (s *postRepository) Close() error {
	var errorOccured error
	if err := s.persist.Close(); err != nil {
		errorOccured = err
	}
	if err := s.selectByUserID.Close(); err != nil {
		errorOccured = err
	}
	if err := s.getByTitle.Close(); err != nil {
		errorOccured = err
	}
	if err := s.getByUID.Close(); err != nil {
		errorOccured = err
	}
	return errorOccured
}

func (s *postRepository) SelectByUserID(ctx context.Context, id int) ([]*model.Post, error) {
	posts := []*model.Post{}
	err := s.selectByUserID.SelectContext(ctx, posts, id)
	return posts, err
}

func (s *postRepository) GetByTitle(ctx context.Context, title string) (*model.Post, error) {
	post := &model.Post{}
	err := s.getByTitle.GetContext(ctx, post, title)
	return post, err
}

func (c *postRepository) Persist(ctx context.Context, post *model.Post) error {
	_, err := c.persist.ExecContext(ctx, post.Title, post.UserID, post.Path,
		post.ParentUID, post.Method, post.Type, post.UniqueID)
	return err
}

func (s *postRepository) GetByUid(ctx context.Context, uid string) (*model.Post, error) {
	post := &model.Post{}
	err := s.getByUID.GetContext(ctx, post, uid)
	return post, err
}
