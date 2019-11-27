package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"gitlab.com/innoserver/pkg/model"
)

const (
	dqlGetByUsername = `SELECT * FROM users WHERE name = ? LIMIT 1`
	persistUser      = `INSERT INTO users (name, email, imei, password)
				   VALUES(?, ?, ?, ?)`
)

type userRepository struct {
	getByUsername   *sqlx.Stmt
	stmtPersistUser *sqlx.Stmt
}

func NewUserRepository(db *sqlx.DB) (*userRepository, error) {
	ctx := context.Background()

	ctxGetByUsername, err := db.PreparexContext(ctx, dqlGetByUsername)
	if err != nil {
		return nil, err
	}

	ctxPersistUser, err := db.PreparexContext(ctx, persistUser)
	if err != nil {
		return nil, err
	}

	return &userRepository{
		getByUsername:   ctxGetByUsername,
		stmtPersistUser: ctxPersistUser,
	}, err
}

func (s *userRepository) Close() error {
	var errorOccured error
	if err := s.stmtPersistUser.Close(); err != nil {
		errorOccured = err
	}
	if err := s.getByUsername.Close(); err != nil {
		errorOccured = err
	}
	return errorOccured
}

func (s *userRepository) GetByUsername(ctx context.Context, name string) (*model.User, error) {
	user := &model.User{}
	err := s.getByUsername.GetContext(ctx, user, name)
	return user, err
}

func (s *userRepository) Persist(ctx context.Context, user *model.User) error {
	_, err := s.stmtPersistUser.ExecContext(ctx, user.Name, user.Email, user.Imei, user.Password)
	return err
}
