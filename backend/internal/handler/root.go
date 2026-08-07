package handler

import (
	"github.com/jackjustus/credittracker/backend/internal/data"
)

type Server struct {
	dao *data.DAO
}

func NewServer(dao *data.DAO) Server {
	return Server{dao}
}
