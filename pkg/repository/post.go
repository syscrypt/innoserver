package repository

import (
	"context"
	"database/sql"

	"github.com/ido50/sqlz"
	"github.com/jmoiron/sqlx"

	"gitlab.com/innoserver/pkg/model"
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
	limit := " LIMIT ?"
	listAllByUserID, _ := sqlz.Newx(db).Select("*").From("posts").
		Where(sqlz.Eq("user_id", "?")).ToSQL(false)

	selectLatest, _ := sqlz.Newx(db).Select("*").From("posts").
		Where(sqlz.IsNull("parent_id"), sqlz.IsNull("group_id")).
		OrderBy(sqlz.Desc("created_at")).ToSQL(false)

	selectLatestOfGroup, _ := sqlz.Newx(db).Select("*").From("posts").
		Where(sqlz.IsNull("parent_id"), sqlz.Eq("group_id", "?")).
		OrderBy(sqlz.Desc("created_at")).ToSQL(false)

	getByTitle, _ := sqlz.Newx(db).Select("*").From("posts").
		Where(sqlz.Like("title", "%?%")).ToSQL(false)

	getByUid, _ := sqlz.Newx(db).Select("*").From("posts").
		Where(sqlz.Eq("unique_id", "?")).ToSQL(false)

	selectChildren, _ := sqlz.Newx(db).Select("*").From("posts").
		Where(sqlz.Eq("parent_id", "?")).OrderBy(sqlz.Desc("created_at")).ToSQL(false)

	persist, _ := sqlz.Newx(db).InsertInto("posts").Columns("title", "user_id", "path",
		"parent_id", "method", "type", "unique_id", "group_id").
		Values("?", "?", "?", "?", "?", "?", "?", "?").ToSQL(false)

	ctxPersistPost, err := db.PreparexContext(ctx, persist)
	if err != nil {
		return nil, err
	}
	ctxSelectByUserID, err := db.PreparexContext(ctx, listAllByUserID)
	if err != nil {
		return nil, err
	}
	ctxGetByTitle, err := db.PreparexContext(ctx, getByTitle)
	if err != nil {
		return nil, err
	}
	ctxGetByUID, err := db.PreparexContext(ctx, getByUid)
	if err != nil {
		return nil, err
	}
	ctxSelectByParent, err := db.PreparexContext(ctx, selectChildren)
	if err != nil {
		return nil, err
	}
	ctxSelectLatest, err := db.PreparexContext(ctx, selectLatest+limit)
	if err != nil {
		return nil, err
	}
	ctxSelectLatestOfGroup, err := db.PreparexContext(ctx, selectLatestOfGroup+limit)
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

func (s *postRepository) GetByTitle(ctx context.Context, title string) ([]*model.Post, error) {
	posts := []*model.Post{}
	err := s.getByTitle.SelectContext(ctx, posts, title)
	return posts, err
}

func (c *postRepository) Persist(ctx context.Context, post *model.Post) error {
	_, err := c.persist.ExecContext(ctx, post.Title, post.UserID, post.Path,
		post.ParentID, post.Method, post.Type, post.UniqueID, post.GroupID)
	return err
}

func (s *postRepository) GetByUid(ctx context.Context, uid string) (*model.Post, error) {
	post := &model.Post{}
	err := s.getByUID.GetContext(ctx, post, uid)
	return post, err
}

func (s *postRepository) SelectByParent(ctx context.Context, parent *model.Post) ([]*model.Post, error) {
	posts := []*model.Post{}
	err := s.selectByParent.SelectContext(ctx, &posts, parent.ID)
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
