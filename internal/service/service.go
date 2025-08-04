package service

import "github.com/worldline-go/saz/internal/database"

type Service struct {
	db *database.Database
}

func New(db *database.Database) *Service {
	return &Service{
		db: db,
	}
}
