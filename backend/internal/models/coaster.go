package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackjustus/credittracker/backend/internal/gen/db"
	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
)

type Coaster struct {
	id             uuid.UUID
	parkID         uuid.UUID
	name           string
	manufacturedAt time.Time
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

func NewCoaster(coaster db.Coaster) *Coaster {
	return &Coaster{
		id:             coaster.ID,
		parkID:         coaster.ParkID,
		name:           coaster.Name,
		manufacturedAt: coaster.ManufacturedAt.Time,
	}
}
