package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"gitlab.com/innoserver/pkg/model"
)

const (
	persistGroup = `INSERT INTO groups (unique_id, title, admin_id, public)
					VALUES (?, ?, ?, ?)`
	dqlGetGroupByUid  = `SELECT * FROM groups WHERE unique_id = ? LIMIT 1`
	sqlAddUserToGroup = `INSERT INTO group_user (group_id, user_id)
						 VALUES (?, ?)`
	sqlGetUsersInGroup = `SELECT u.name, u.email, u.imei FROM users u WHERE
						  u.id IN (SELECT gu.user_id FROM group_user gu WHERE
						  gu.group_id = ?)`
	sqlGetUserIDsInGrp = `SELECT u.id FROM users u WHERE u.id IN
						  (SELECT gu.user_id FROM group_user gu WHERE
						  gu.group_id =?)`
	sqlUpdateGroup = `UPDATE groups SET public = ? WHERE id = ?`
)

type groupRepository struct {
	persistGroup         *sqlx.Stmt
	getByUid             *sqlx.Stmt
	stmtAddUserToGroup   *sqlx.Stmt
	stmtGetUsersInGroup  *sqlx.Stmt
	stmtGetUserIDsInGrp  *sqlx.Stmt
	stmtUpdateVisibility *sqlx.Stmt
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
	if err != nil {
		return nil, err
	}
	ctxGetUsersInGroup, err := db.PreparexContext(ctx, sqlGetUsersInGroup)
	if err != nil {
		return nil, err
	}
	ctxGetUserIDsInGrp, err := db.PreparexContext(ctx, sqlGetUserIDsInGrp)
	if err != nil {
		return nil, err
	}
	ctxUpdateVisibility, err := db.PreparexContext(ctx, sqlUpdateGroup)
	if err != nil {
		return nil, err
	}
	return &groupRepository{
		persistGroup:         ctxPersist,
		getByUid:             ctxGetByUid,
		stmtAddUserToGroup:   ctxAddUserToGroup,
		stmtGetUsersInGroup:  ctxGetUsersInGroup,
		stmtGetUserIDsInGrp:  ctxGetUserIDsInGrp,
		stmtUpdateVisibility: ctxUpdateVisibility,
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
	if err := s.stmtGetUsersInGroup.Close(); err != nil {
		errorOccured = err
	}
	if err := s.stmtGetUserIDsInGrp.Close(); err != nil {
		errorOccured = err
	}
	if err := s.stmtUpdateVisibility.Close(); err != nil {
		errorOccured = err
	}
	return errorOccured
}

func (s *groupRepository) Persist(ctx context.Context, group *model.Group) error {
	_, err := s.persistGroup.ExecContext(ctx, group.UniqueID, group.Title,
		group.AdminID, group.Public)
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
	_, err := s.stmtAddUserToGroup.ExecContext(ctx, group.ID, user.ID)
	return err
}

func (s *groupRepository) GetUsersInGroup(ctx context.Context, group *model.Group) ([]*model.User, error) {
	users := []*model.User{}
	err := s.stmtGetUsersInGroup.SelectContext(ctx, &users, group.ID)
	return users, err
}

func (s *groupRepository) GetUserIDsInGroup(ctx context.Context, group *model.Group) ([]*model.User, error) {
	users := []*model.User{}
	err := s.stmtGetUserIDsInGrp.SelectContext(ctx, &users, group.ID)
	return users, err
}

func (s *groupRepository) IsUserInGroup(ctx context.Context, user *model.User, group *model.Group) (bool, error) {
	users, err := s.GetUserIDsInGroup(ctx, group)
	if err != nil {
		return true, err
	}
	for _, u := range users {
		if u.ID == user.ID {
			return true, nil
		}
	}
	return false, nil
}

func (s *groupRepository) UpdateVisibility(ctx context.Context, group *model.Group) error {
	_, err := s.stmtUpdateVisibility.ExecContext(ctx, group.Public, group.ID)
	return err
}
