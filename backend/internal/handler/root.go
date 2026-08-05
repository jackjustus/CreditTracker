package handler

import "github.com/jackjustus/credittracker/backend/internal/gen/db"

type Server struct {
	dbc *db.Queries
}

func NewServer(dbc *db.Queries) Server {
	return Server{dbc}
}
