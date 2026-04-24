package flagapi

import "database/sql"

type Server struct {
	db      *sql.DB
	hashKey []byte
}

func New(db *sql.DB, hashSecret string) *Server {
	return &Server{
		db:      db,
		hashKey: []byte(hashSecret),
	}
}
