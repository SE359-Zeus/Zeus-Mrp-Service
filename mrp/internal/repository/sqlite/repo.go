package sqlite

import "gorm.io/gorm"

type Repository = sqliteMRPRepository

func New(db *gorm.DB) *Repository {
	return &sqliteMRPRepository{db: db}
}
