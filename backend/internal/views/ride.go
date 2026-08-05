package views

import (
	"github.com/jackjustus/credittracker/backend/internal/gen/db"
	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/life4/genesis/slices"
)

type Ride struct {
	coaster db.Coaster
	time    pgtype.Timestamptz
}

func (e *Ride) ToCredit() api.Credit {
	return api.Credit{Name: e.coaster.Name}
}

type Rides []*Ride

func (es Rides) ToCredits() api.Credits {
	return slices.Map(es, func(event *Ride) api.Credit {
		return event.ToCredit()
	})
}

func NewRideEvents(rows []db.GetRideEventsRow) Rides {
	return slices.Map(rows, func(row db.GetRideEventsRow) *Ride {
		return NewRideEvent(row)
	})
}

func NewRideEvent(row db.GetRideEventsRow) *Ride {
	return &Ride{
		coaster: row.Coaster,
		time:    row.RideEvent.CreatedAt,
	}
}
