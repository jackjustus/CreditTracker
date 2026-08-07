package db

import (
	"github.com/google/uuid"
	"github.com/life4/genesis/slices"
)

type GetRideEventsRows []GetRideEventsRow

func (rows GetRideEventsRows) ByID() map[uuid.UUID]GetRideEventsRows {
	return slices.GroupBy(rows, func(row GetRideEventsRow) uuid.UUID {
		return row.Coaster.ID
	})
}

func (rows GetRideEventsRows) IsSingleCoaster() bool {
	return slices.All(rows, func(row GetRideEventsRow) bool {
		return row.Coaster.ID == rows[0].Coaster.ID
	})
}
