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

func (c *Coaster) ToAPI() api.Coaster {
	return api.Coaster{
		Id:             c.id,
		ManufacturedAt: c.manufacturedAt,
		Name:           c.name,
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

func (cs Coasters) ToAPI() api.Coasters {
	return slices.Map(cs, func(c *Coaster) api.Coaster {
		return c.ToAPI()
	})
}

type CoasterWithPark struct {
	coaster *Coaster
	park    *Park
}

type CoastersWithPark []*CoasterWithPark

func (cwp *CoasterWithPark) Coaster() *Coaster {
	return cwp.coaster
}

func (cwp *CoasterWithPark) Park() *Park {
	return cwp.park
}

func (cwp *CoasterWithPark) ToAPI() api.Coaster {
	return api.Coaster{
		Id:             cwp.coaster.id,
		ManufacturedAt: cwp.coaster.manufacturedAt,
		Name:           cwp.coaster.name,
		Park: api.Park{
			City:    cwp.park.city,
			Country: cwp.park.country,
			Id:      cwp.park.id,
			Name:    cwp.park.name,
		},
	}
}

func (cpws CoastersWithPark) ToAPI() api.Coasters {
	return slices.Map(cpws, func(cwp *CoasterWithPark) api.Coaster {
		return cwp.ToAPI()
	})
}

func NewCoasterWithParkFromDB(coaster db.Coaster) *CoasterWithPark {
	return &CoasterWithPark{
		coaster: NewCoaster(coaster),
		park: &Park{
			id: coaster.ParkID,
		},
	}
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
