package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"gitlab.com/innoserver/pkg/model"
)

const (
	persistGroup = `INSERT INTO groups (unique_id, title, admin_id)
					VALUES (?, ?, ?)`
)

type groupRepository struct {
	persistGroup *sqlx.Stmt
}

func NewGroupRepository(db *sqlx.DB) (*groupRepository, error) {
	ctx := context.Background()
	ctxPersist, err := db.PreparexContext(ctx, persistGroup)
	if err != nil {
		return nil, err
	}
	return &groupRepository{
		persistGroup: ctxPersist,
	}, err
}

func (s *groupRepository) Close() error {
	var errorOccured error
	if err := s.persistGroup.Close(); err != nil {
		errorOccured = err
	}
	return errorOccured
}

func (s *groupRepository) Persist(ctx context.Context, group *model.Group) error {
	_, err := s.persistGroup.ExecContext(ctx, group.UniqueID, group.Title, group.AdminID)
	return err
}
