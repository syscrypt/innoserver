package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"gitlab.com/innoserver/pkg/model"
)

const (
	dqlAllPostsByUserID = `SELECT * FROM posts WHERE user_id = ?`
	selectLatestPosts   = `SELECT * FROM posts WHERE parent_uid = ""
						   AND group_id IS NULL ORDER BY created_at DESC LIMIT ?`
	selectLatestPostsOfGroup = `SELECT * FROM posts WHERE parent_uid = ""
								AND group_id = ? ORDER BY created_at DESC LIMIT ?`
	dqlGetPostByTitle     = `SELECT * FROM posts WHERE title = ?`
	dqlGetPostByUid       = `SELECT * FROM posts WHERE unique_id = ? LIMIT 1`
	selectChildPostsByUid = `SELECT * FROM posts WHERE parent_uid = ? ORDER BY created_at DESC`
	persistPost           = `INSERT INTO posts
						   (title, user_id, path, parent_uid,
							method, type, unique_id, group_id)
							VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
)

type postRepository struct {
	persist             *sqlx.Stmt
	selectByUserID      *sqlx.Stmt
	getByTitle          *sqlx.Stmt
	getByUID            *sqlx.Stmt
	selectByParent      *sqlx.Stmt
	selectLatest        *sqlx.Stmt
	selectLatestOfGroup *sqlx.Stmt
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
	ctxSelectByParent, err := db.PreparexContext(ctx, selectChildPostsByUid)
	if err != nil {
		return nil, err
	}
	ctxSelectLatest, err := db.PreparexContext(ctx, selectLatestPosts)
	if err != nil {
		return nil, err
	}
	ctxSelectLatestOfGroup, err := db.PreparexContext(ctx, selectLatestPostsOfGroup)
	if err != nil {
		return nil, err
	}

	return &postRepository{
		persist:             ctxPersistPost,
		selectByUserID:      ctxSelectByUserID,
		getByTitle:          ctxGetByTitle,
		getByUID:            ctxGetByUID,
		selectByParent:      ctxSelectByParent,
		selectLatest:        ctxSelectLatest,
		selectLatestOfGroup: ctxSelectLatestOfGroup,
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
	if err := s.selectByParent.Close(); err != nil {
		errorOccured = err
	}
	if err := s.selectLatest.Close(); err != nil {
		errorOccured = err
	}
	if err := s.selectLatestOfGroup.Close(); err != nil {
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
		post.ParentUID, post.Method, post.Type, post.UniqueID, post.GroupID)
	return err
}

func (s *postRepository) GetByUid(ctx context.Context, uid string) (*model.Post, error) {
	post := &model.Post{}
	err := s.getByUID.GetContext(ctx, post, uid)
	return post, err
}

func (s *postRepository) SelectByParentUid(ctx context.Context, uid string) ([]*model.Post, error) {
	posts := []*model.Post{}
	err := s.selectByParent.SelectContext(ctx, &posts, uid)
	return posts, err
}

func (s *postRepository) UniqueIdExists(ctx context.Context, uid string) (bool, error) {
	if _, err := s.GetByUid(ctx, uid); err != nil && err != sql.ErrNoRows {
		return true, err
	}
	return false, nil
}

func (s *postRepository) SelectLatest(ctx context.Context, limit uint64) ([]*model.Post, error) {
	posts := []*model.Post{}
	err := s.selectLatest.SelectContext(ctx, &posts, limit)
	return posts, err
}

func (s *postRepository) SelectLatestOfGroup(ctx context.Context, group *model.Group, limit uint64) ([]*model.Post, error) {
	posts := []*model.Post{}
	err := s.selectLatestOfGroup.SelectContext(ctx, &posts, group.ID, limit)
	return posts, err
}
