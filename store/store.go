package store

import "database/sql"

type store struct {
	db *sql.DB
}

func New(db *sql.DB) *store {
	return &store{db: db}
}
