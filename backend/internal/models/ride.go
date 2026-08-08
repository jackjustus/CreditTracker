package models

import (
	"time"

	"github.com/jackjustus/credittracker/backend/internal/gen/db"
	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
	"github.com/life4/genesis/slices"
)

type Ride struct {
	coaster  *Coaster
	riddenAt time.Time
}

func (r *Ride) ToAPI() api.Ride {
	return r.coaster.ToRide(r.riddenAt)
}

func NewRide(coaster *Coaster, riddenAt time.Time) *Ride {
	return &Ride{
		coaster:  coaster,
		riddenAt: riddenAt,
	}
}

type Rides []*Ride

func (rds Rides) ToAPI() api.Rides {
	return slices.Map(rds, func(ride *Ride) api.Ride {
		return ride.ToAPI()
	})
}

func NewRides(rows db.GetRideEventsRows) Rides {
	rides := slices.Map(rows, func(row db.GetRideEventsRow) *Ride {
		return NewRide(
			NewCoaster(row.Coaster),
			row.RideEvent.CreatedAt.Time,
		)
	})
	return rides
}
