package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"gitlab.com/innoserver/pkg/model"
)

const (
	dqlAllPostsByUserID = `SELECT * FROM posts WHERE user_id = ?`
	dqlGetPostByTitle   = `SELECT * FROM posts WHERE title = ?`
	persistPost         = `INSERT INTO posts
						   (title, user_id, path, parent_uid,
							method, type, unique_id)
						   VALUES (?, ?, ?, ?, ?, ?, ?)`
)

type postRepository struct {
	persist        *sqlx.Stmt
	selectByUserID *sqlx.Stmt
	getByTitle     *sqlx.Stmt
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

	return &postRepository{
		persist:        ctxPersistPost,
		selectByUserID: ctxSelectByUserID,
		getByTitle:     ctxGetByTitle,
	}, err
}

func (s *postRepository) SelectByUserID(ctx context.Context, id int) ([]*model.Post, error) {
	posts := []*model.Post{}
	err := s.selectByUserID.SelectContext(ctx, posts, id)
	return posts, err
}

func (s *postRepository) GetByTitle(ctx context.Context, title string) (*model.Post, error) {
	post := &model.Post{}
	err := s.getByTitle.SelectContext(ctx, post, title)
	return post, err
}

func (c *postRepository) Persist(ctx context.Context, post *model.Post) error {
	_, err := c.persist.ExecContext(ctx, post.Title, post.UserID, post.Path,
		post.ParentUID, post.Method, post.Type, post.UniqueID)
	return err
}
