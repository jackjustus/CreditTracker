package views

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackjustus/credittracker/backend/internal/gen/db"
	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
)

type Coaster struct {
	ID   uuid.UUID
	Name string
}

func (c *Coaster) ToRide(time time.Time) api.Ride {
	return api.Ride{
		Coaster: api.Coaster{
			Id:   c.ID,
			Name: c.Name,
		},
		RiddenAt: time,
	}
}

func NewCoaster(row db.HydrateCoasterRow) *Coaster {
	return &Coaster{
		row.Coaster.ID,
		row.Coaster.Name,
	}
}
