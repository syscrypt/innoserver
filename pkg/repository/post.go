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
	getByTitleInGroup   *sqlx.Stmt
	getByUID            *sqlx.Stmt
	selectByParent      *sqlx.Stmt
	selectLatest        *sqlx.Stmt
	selectLatestOfGroup *sqlx.Stmt
	addOptions          *sqlx.Stmt
	removeOptions       *sqlx.Stmt
	selectOptions       *sqlx.Stmt
}

func NewPostRepository(db *sqlx.DB) (*postRepository, error) {
	ctx := context.Background()
	limit := " LIMIT ?"
	listAllByUserID, _ := sqlz.Newx(db).Select("*").From("posts").
		Where(sqlz.Eq("user_id", "?")).ToSQL(false)

	selectLatest, _ := sqlz.Newx(db).Select("*").From("detailed_posts").
		Where(sqlz.IsNull("parent_id"), sqlz.IsNull("group_id")).
		OrderBy(sqlz.Desc("created_at")).ToSQL(false)

	selectLatestOfGroup, _ := sqlz.Newx(db).Select("*").From("detailed_posts").
		Where(sqlz.IsNull("parent_id"), sqlz.Eq("group_id", "?")).
		OrderBy(sqlz.Desc("created_at")).ToSQL(false)

	getByTitle, _ := sqlz.Newx(db).Select("*").From("detailed_posts").
		Where(sqlz.IsNull("parent_id"), sqlz.IsNull("group_id")).
		Where(sqlz.Like("title", "%?%")).
		OrderBy(sqlz.Desc("created_at")).ToSQL(false)

	getByTitleInGroup, _ := sqlz.Newx(db).Select("*").From("detailed_posts").
		Where(sqlz.IsNull("parent_id")).
		Where(sqlz.Like("title", "%?%")).
		Where(sqlz.Eq("group_id", "?")).
		OrderBy(sqlz.Desc("created_at")).ToSQL(false)

	getByUid, _ := sqlz.Newx(db).Select("*").From("detailed_posts").
		Where(sqlz.Eq("unique_id", "?")).ToSQL(false)

	selectChildren, _ := sqlz.Newx(db).Select("*").From("detailed_posts").
		Where(sqlz.Eq("parent_id", "?")).OrderBy(sqlz.Desc("created_at")).ToSQL(false)

	persist, _ := sqlz.Newx(db).InsertInto("posts").Columns("title", "user_id", "path",
		"parent_id", "method", "type", "unique_id", "group_id").
		Values("?", "?", "?", "?", "?", "?", "?", "?").ToSQL(false)

	addOptions, _ := sqlz.Newx(db).InsertInto("options").Columns("post_uid",
		"opt_key", "opt_value").Values("?", "?", "?").ToSQL(false)

	removeOptions, _ := sqlz.Newx(db).DeleteFrom("options").Where(
		sqlz.Eq("post_uid", "?")).ToSQL(false)

	selectOptions, _ := sqlz.Newx(db).Select("*").From("options").
		Where(sqlz.Eq("post_uid", "?")).ToSQL(false)

	ctxPersistPost, err := db.PreparexContext(ctx, persist)
	if err != nil {
		return nil, err
	}
	ctxSelectByUserID, err := db.PreparexContext(ctx, listAllByUserID)
	if err != nil {
		return nil, err
	}
	ctxGetByTitle, err := db.PreparexContext(ctx, getByTitle+limit)
	if err != nil {
		return nil, err
	}
	ctxGetByTitleInGroup, err := db.PreparexContext(ctx, getByTitleInGroup+limit)
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
	ctxAddOptions, err := db.PreparexContext(ctx, addOptions)
	if err != nil {
		return nil, err
	}
	ctxRemoveOptions, err := db.PreparexContext(ctx, removeOptions)
	if err != nil {
		return nil, err
	}
	ctxSelectOptions, err := db.PreparexContext(ctx, selectOptions)
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
		addOptions:          ctxAddOptions,
		removeOptions:       ctxRemoveOptions,
		selectOptions:       ctxSelectOptions,
		getByTitleInGroup:   ctxGetByTitleInGroup,
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
	if err := s.getByTitleInGroup.Close(); err != nil {
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
	if err := s.addOptions.Close(); err != nil {
		errorOccured = err
	}
	if err := s.removeOptions.Close(); err != nil {
		errorOccured = err
	}
	if err := s.selectOptions.Close(); err != nil {
		errorOccured = err
	}
	return errorOccured
}

func (s *postRepository) SelectByUserID(ctx context.Context, id int) ([]*model.Post, error) {
	posts := []*model.Post{}
	err := s.selectByUserID.SelectContext(ctx, posts, id)
	if err == nil {
		err = s.appendOptionsMult(ctx, posts)
	}
	return posts, err
}

func (s *postRepository) GetByTitle(ctx context.Context, title string, limit int64) ([]*model.Post, error) {
	posts := []*model.Post{}
	err := s.getByTitle.SelectContext(ctx, &posts, "%"+title+"%", limit)
	if err == nil {
		err = s.appendOptionsMult(ctx, posts)
	}
	return posts, err
}

func (s *postRepository) GetByTitleInGroup(ctx context.Context, title string,
	group *model.Group, limit int64) ([]*model.Post, error) {
	posts := []*model.Post{}
	err := s.getByTitleInGroup.SelectContext(ctx, &posts, "%"+title+"%", group.ID, limit)
	if err == nil {
		err = s.appendOptionsMult(ctx, posts)
	}
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
	if err == nil {
		err = s.appendOptions(ctx, post)
	}
	return post, err
}

func (s *postRepository) SelectByParent(ctx context.Context, parent *model.Post) ([]*model.Post, error) {
	posts := []*model.Post{}
	err := s.selectByParent.SelectContext(ctx, &posts, parent.ID)
	if err == nil {
		err = s.appendOptionsMult(ctx, posts)
	}
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
	if err == nil {
		err = s.appendOptionsMult(ctx, posts)
	}
	return posts, err
}

func (s *postRepository) SelectLatestOfGroup(ctx context.Context, group *model.Group, limit uint64) ([]*model.Post, error) {
	posts := []*model.Post{}
	err := s.selectLatestOfGroup.SelectContext(ctx, &posts, group.ID, limit)
	if err == nil {
		err = s.appendOptionsMult(ctx, posts)
	}
	return posts, err
}

func (s *postRepository) appendOptions(ctx context.Context, post *model.Post) error {
	options, err := s.SelectOptions(ctx, post)
	post.Options = append(post.Options, options...)
	return err
}

func (s *postRepository) appendOptionsMult(ctx context.Context, posts []*model.Post) error {
	for _, post := range posts {
		err := s.appendOptions(ctx, post)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *postRepository) AddOptions(ctx context.Context, post *model.Post, options []*model.Option) error {
	for _, v := range options {
		_, err := s.addOptions.ExecContext(ctx, post.UniqueID, v.Key, v.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *postRepository) RemoveOptions(ctx context.Context, post *model.Post) error {
	_, err := s.removeOptions.ExecContext(ctx, post.UniqueID)
	return err
}

func (s *postRepository) SetOptions(ctx context.Context, post *model.Post, options []*model.Option) error {
	err := s.RemoveOptions(ctx, post)
	if err != nil {
		return err
	}
	return s.AddOptions(ctx, post, options)
}

func (s *postRepository) SelectOptions(ctx context.Context, post *model.Post) ([]*model.Option, error) {
	options := []*model.Option{}
	err := s.selectOptions.SelectContext(ctx, &options, post.UniqueID)
	return options, err
}
