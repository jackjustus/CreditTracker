package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackjustus/credittracker/backend/internal/gen/db"
	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
	"github.com/life4/genesis/slices"
)

type Coaster struct {
	id             uuid.UUID
	parkID         uuid.UUID
	name           string
	manufacturedAt time.Time
}

func (c *Coaster) ID() uuid.UUID {
	return c.id
}

func (c *Coaster) Name() string {
	return c.name
}

func (c *Coaster) APIRef() api.CoasterRef {
	return api.CoasterRef{
		Id: c.id,
	}
}

func (c *Coaster) ToRide(time time.Time) api.Ride {
	return api.Ride{
		Coaster: api.CoasterRef{
			Id: c.id,
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

type Coasters []*Coaster

func (cs Coasters) ToAPI() api.CoasterRefs {
	return slices.Map(cs, func(c *Coaster) api.CoasterRef {
		return c.APIRef()
	})
}

type CoasterWithPark struct {
	coaster *Coaster
	park    *Park
}

func (c *CoasterWithPark) Coaster() *Coaster {
	return c.coaster
}

func (c *CoasterWithPark) Park() *Park {
	return c.park
}

func NewCoasterWithPark(coaster *Coaster, park *Park) (*CoasterWithPark, error) {
	if coaster.parkID != park.id {
		return nil, errors.New("cannot create coaster and park without relationship")
	}
	return &CoasterWithPark{
		coaster: coaster,
		park:    park,
	}, nil
}
