package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"gitlab.com/innoserver/pkg/model"
)

const (
	dqlGetByUsername = `SELECT * FROM users WHERE name = ?`
)

type userRepository struct {
	getByUsername *sqlx.Stmt
}

func NewUserRepository(db *sqlx.DB) (*userRepository, error) {
	ctx := context.Background()

	ctxGetByUsername, err := db.PreparexContext(ctx, dqlGetByUsername)
	if err != nil {
		return nil, err
	}

	return &userRepository{
		getByUsername: ctxGetByUsername,
	}, err
}

func (s *userRepository) GetByUsername(ctx context.Context, name string) (*model.User, error) {
	user := &model.User{}
	err := s.getByUsername.GetContext(ctx, &user, name)
	return user, err
}
