package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"gitlab.com/innoserver/pkg/model"
)

const (
	persistGroup = `INSERT INTO groups (unique_id, title, admin_id)
					VALUES (?, ?, ?)`
	dqlGetGroupByUid  = `SELECT * FROM groups WHERE unique_id = ? LIMIT 1`
	sqlAddUserToGroup = `INSERT INTO group_user (group_id, user_id)
						 VALUES (?, ?)`
)

type groupRepository struct {
	persistGroup       *sqlx.Stmt
	getByUid           *sqlx.Stmt
	stmtAddUserToGroup *sqlx.Stmt
}

func NewGroupRepository(db *sqlx.DB) (*groupRepository, error) {
	ctx := context.Background()
	ctxPersist, err := db.PreparexContext(ctx, persistGroup)
	if err != nil {
		return nil, err
	}
	ctxGetByUid, err := db.PreparexContext(ctx, dqlGetGroupByUid)
	if err != nil {
		return nil, err
	}
	ctxAddUserToGroup, err := db.PreparexContext(ctx, sqlAddUserToGroup)
	return &groupRepository{
		persistGroup:       ctxPersist,
		getByUid:           ctxGetByUid,
		stmtAddUserToGroup: ctxAddUserToGroup,
	}, err
}

func (s *groupRepository) Close() error {
	var errorOccured error
	if err := s.persistGroup.Close(); err != nil {
		errorOccured = err
	}
	if err := s.getByUid.Close(); err != nil {
		errorOccured = err
	}
	if err := s.stmtAddUserToGroup.Close(); err != nil {
		errorOccured = err
	}
	return errorOccured
}

func (s *groupRepository) Persist(ctx context.Context, group *model.Group) error {
	_, err := s.persistGroup.ExecContext(ctx, group.UniqueID, group.Title, group.AdminID)
	return err
}

func (s *groupRepository) GetByUid(ctx context.Context, uid string) (*model.Group, error) {
	group := &model.Group{}
	err := s.getByUid.GetContext(ctx, group, uid)
	return group, err
}

func (s *groupRepository) UniqueIdExists(ctx context.Context, uid string) (bool, error) {
	if _, err := s.GetByUid(ctx, uid); err != nil && err != sql.ErrNoRows {
		return true, err
	}
	return false, nil
}

func (s *groupRepository) AddUserToGroup(ctx context.Context, user *model.User, group *model.Group) error {
	_, err := s.stmtAddUserToGroup.ExecContext(ctx, user.ID, group.ID)
	return err
}
