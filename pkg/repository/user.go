package repository

import (
	"context"

	"github.com/ido50/sqlz"
	"github.com/jmoiron/sqlx"

	"gitlab.com/innoserver/pkg/model"
)

type userRepository struct {
	getByUsername   *sqlx.Stmt
	getByEmail      *sqlx.Stmt
	stmtPersistUser *sqlx.Stmt
}

func NewUserRepository(db *sqlx.DB) (*userRepository, error) {
	ctx := context.Background()
	GetByName, _ := sqlz.Newx(db).Select("*").From("users").
		Where(sqlz.Eq("name", "?")).Limit(1).ToSQL(false)
	GetByMail, _ := sqlz.Newx(db).Select("*").From("users").
		Where(sqlz.Eq("email", "?")).Limit(1).ToSQL(false)
	Persist, _ := sqlz.Newx(db).InsertInto("users").Columns("name", "email",
		"imei", "password").Values("?", "?", "?", "?").ToSQL(false)

	ctxGetByUsername, err := db.PreparexContext(ctx, GetByName)
	if err != nil {
		return nil, err
	}
	ctxPersistUser, err := db.PreparexContext(ctx, Persist)
	if err != nil {
		return nil, err
	}
	ctxByEmail, err := db.PreparexContext(ctx, GetByMail)
	if err != nil {
		return nil, err
	}
	return &userRepository{
		getByUsername:   ctxGetByUsername,
		stmtPersistUser: ctxPersistUser,
		getByEmail:      ctxByEmail,
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
	if err := s.getByEmail.Close(); err != nil {
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

func (s *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := s.getByEmail.GetContext(ctx, user, email)
	return user, err
}
