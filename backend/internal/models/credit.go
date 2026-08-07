package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackjustus/credittracker/backend/internal/gen/db"
	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
	"github.com/life4/genesis/slices"
)

type credit struct {
	coasterID     uuid.UUID
	coasterName   string
	firstRiddenAt time.Time
	lastRiddenAt  time.Time
	rideCount     int
}

func newCredit(events db.GetRideEventsRows) (*credit, error) {
	numRides := len(events)
	if numRides < 1 {
		return nil, errors.New("must have at least one ride event to construct a credit")
	}

	if !events.IsSingleCoaster() {
		return nil, errors.New("a single credit must be constructed from a single coaster")
	}

	sortedEvents := slices.SortBy(events, func(event db.GetRideEventsRow) int64 {
		return event.RideEvent.CreatedAt.Time.Unix()
	})
	firstRide := sortedEvents[0]
	lastRide := sortedEvents[len(sortedEvents)-1]

	return &credit{
		coasterID:     firstRide.Coaster.ID,
		coasterName:   firstRide.Coaster.Name,
		firstRiddenAt: firstRide.RideEvent.CreatedAt.Time,
		lastRiddenAt:  lastRide.RideEvent.CreatedAt.Time,
		rideCount:     numRides,
	}, nil
}

func (crd *credit) ToAPI() api.Credit {
	return api.Credit{
		Id:            crd.coasterID,
		Name:          crd.coasterName,
		FirstRiddenAt: crd.firstRiddenAt,
		LastRiddenAt:  crd.lastRiddenAt,
		RideCount:     crd.rideCount,
	}
}

type Credits []*credit

func (crds Credits) ToAPI() api.Credits {
	return slices.Map(crds, func(crd *credit) api.Credit {
		return crd.ToAPI()
	})
}

// NewCredits groups ride event rows by coaster and populates a credit for each.
func NewCredits(events db.GetRideEventsRows) (*Credits, error) {
	eventsByID := events.ByID()
	creds := make(Credits, 0, len(eventsByID))

	for _, rows := range eventsByID {
		credit, err := newCredit(rows)
		if err != nil {
			return nil, err
		}
		creds = append(creds, credit)
	}

	return &creds, nil
}
