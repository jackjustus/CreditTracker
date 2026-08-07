package models

import (
	"time"

	"github.com/google/uuid"
)

type Ride struct {
	coasterID uuid.UUID
	time      time.Time
}

//func (e *Ride) toCredit() api.Credit {
//	return api.Credit{
//		FirstRiddenAt: e.time,
//		Name: e.coaster.Name,
//	}
//}

type Rides struct {
	rides []*Ride
}

//func (rides Rides) ToCredits() api.Credits {
//	ridesByID := slices.GroupBy(rides, func(ride *Ride) uint32 {
//		return ride.coaster.ID.ID()
//	})
//	slices.Each(ridesByID, func(coasterID uint32, ridesForCoaster []Rides) {
//
//	})
//
//	return api.Credit{
//		FirstRiddenAt: time.Time{},
//		Id:            api.UUID{},
//		LastRiddenAt:  time.Time{},
//		Name:          "",
//		RideCount:     0,
//	}
//}

//// NewRideEvents creates a ride events obj from the GetRideEvents query.
//func NewRideEvents(rows db.GetRideEventsRows) Rides {
//	rideEventsByID := rows.ByID()
//	for id, rides := range rideEventsByID {
//
//	}
//	slices.Map(rideEventsByID func() api.Credit {
//
//	})
//
//	//return slices.Map(rows, func(row db.GetRideEventsRow) *Ride {
//	//	return NewRideEvent(row)
//	//})
//}
//
//func NewRide(row db.GetRideEventsRow) *Ride {
//	return &Ride{
//		coaster: row.Coaster,
//		time:    row.RideEvent.CreatedAt,
//	}
//}
