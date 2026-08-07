package models

import (
	"time"

	"github.com/google/uuid"
	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
)

type Coaster struct {
	id   uuid.UUID
	name string
}

func (c *Coaster) ToRide(time time.Time) api.Ride {
	return api.Ride{
		Coaster: api.Coaster{
			Id:   c.id,
			Name: c.name,
		},
		RiddenAt: time,
	}
}

func NewCoaster(id uuid.UUID, name string) *Coaster {
	return &Coaster{
		id,
		name,
	}
}
