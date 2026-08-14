package handler

import (
	"github.com/jackjustus/credittracker/backend/internal/data"
	"github.com/jackjustus/credittracker/backend/internal/embed"
)

type Server struct {
	dao      *data.DAO
	embedder *embed.Client
}

func NewServer(dao *data.DAO, embedder *embed.Client) Server {
	return Server{dao, embedder}
}
