package models

import (
	"github.com/google/uuid"
	"github.com/jackjustus/credittracker/backend/internal/gen/db"
)

type Park struct {
	id      uuid.UUID
	name    string
	city    string
	country string
}

func (p *Park) ID() uuid.UUID {
	return p.id
}

func (p *Park) Name() string {
	return p.name
}

func NewPark(row db.Park) *Park {
	return &Park{
		id:      row.ID,
		name:    row.Name,
		city:    row.City,
		country: row.Country,
	}
}
