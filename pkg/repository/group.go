package repository

import (
	"context"
	"database/sql"

	"github.com/ido50/sqlz"
	"github.com/jmoiron/sqlx"

	"gitlab.com/innoserver/pkg/model"
)

type groupRepository struct {
	persistGroup         *sqlx.Stmt
	getByUid             *sqlx.Stmt
	stmtAddUserToGroup   *sqlx.Stmt
	stmtGetUsersInGroup  *sqlx.Stmt
	stmtGetUserIDsInGrp  *sqlx.Stmt
	stmtUpdateVisibility *sqlx.Stmt
	stmtSelectByUser     *sqlx.Stmt
	stmtRemove           *sqlx.Stmt
}

func NewGroupRepository(db *sqlx.DB) (*groupRepository, error) {
	ctx := context.Background()
	persist, _ := sqlz.Newx(db).InsertInto("groups").Columns("unique_id", "title",
		"admin_id", "public").Values("?", "?", "?", "?").ToSQL(false)

	getByUid, _ := sqlz.Newx(db).Select("*").From("groups").
		Where(sqlz.Eq("unique_id", "?")).ToSQL(false)

	addUser, _ := sqlz.Newx(db).InsertInto("group_user").
		Columns("group_id", "user_id").Values("?", "?").ToSQL(false)

	listMembers, _ := sqlz.Newx(db).Select("users.name", "users.email", "users.imei").
		From("users").InnerJoin("group_user", sqlz.Eq("users.id", sqlz.Indirect("group_user.user_id"))).
		Where(sqlz.Eq("group_user.group_id", "?")).ToSQL(false)

	listUserIds, _ := sqlz.Newx(db).Select("user_id AS id").From("group_user").
		Where(sqlz.Eq("group_id", "?")).ToSQL(false)

	updateVisibility, _ := sqlz.Newx(db).Update("groups").Set("public", "?").
		Where(sqlz.Eq("id", "?")).ToSQL(false)

	selectByUser, _ := sqlz.Newx(db).Select("g.*").From("group_user gu").
		LeftJoin("groups g", sqlz.Eq("g.id", sqlz.Indirect("gu.group_id"))).
		Where(sqlz.Eq("gu.user_id", "?")).ToSQL(false)

	remove, _ := sqlz.Newx(db).DeleteFrom("groups").Where(sqlz.Eq("id", "?")).ToSQL(false)

	ctxPersist, err := db.PreparexContext(ctx, persist)
	if err != nil {
		return nil, err
	}
	ctxGetByUid, err := db.PreparexContext(ctx, getByUid)
	if err != nil {
		return nil, err
	}
	ctxAddUserToGroup, err := db.PreparexContext(ctx, addUser)
	if err != nil {
		return nil, err
	}
	ctxGetUsersInGroup, err := db.PreparexContext(ctx, listMembers)
	if err != nil {
		return nil, err
	}
	ctxGetUserIDsInGrp, err := db.PreparexContext(ctx, listUserIds)
	if err != nil {
		return nil, err
	}
	ctxUpdateVisibility, err := db.PreparexContext(ctx, updateVisibility)
	if err != nil {
		return nil, err
	}
	ctxSelectByUser, err := db.PreparexContext(ctx, selectByUser)
	if err != nil {
		return nil, err
	}
	ctxRemoveGroup, err := db.PreparexContext(ctx, remove)
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
		stmtSelectByUser:     ctxSelectByUser,
		stmtRemove:           ctxRemoveGroup,
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
	if err := s.stmtSelectByUser.Close(); err != nil {
		errorOccured = err
	}
	if err := s.stmtRemove.Close(); err != nil {
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

func (s *groupRepository) SelectByUser(ctx context.Context, user *model.User) ([]*model.Group, error) {
	groups := []*model.Group{}
	err := s.stmtSelectByUser.SelectContext(ctx, &groups, user.ID)
	return groups, err
}

func (s *groupRepository) RemoveGroup(ctx context.Context, group *model.Group) error {
	_, err := s.stmtRemove.ExecContext(ctx, group.ID)
	return err
}
