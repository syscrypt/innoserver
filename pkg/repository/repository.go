package repository

import (
	"github.com/jmoiron/sqlx"
)

type repository struct {
}

type Repository interface {
}

func NewService(db *sqlx.DB) *repository {
	return &repository{}
}
