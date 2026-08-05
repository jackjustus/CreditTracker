package handler

import (
	"context"

	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
	"github.com/jackjustus/credittracker/backend/internal/views"
)

func (s Server) ListCredits(ctx context.Context, req api.ListCreditsRequestObject) (api.ListCreditsResponseObject, error) {
	rideEvents, err := s.dbc.GetRideEvents(ctx)
	if err != nil {
		return nil, err
	}
	//coasterNames := slices.Map(rideEvents, func(event db.GetRideEventsRow) string {
	//	return event.CoasterName
	//})
	rides := views.NewRideEvents(rideEvents)
	return api.ListCredits200JSONResponse(rides.ToCredits()), nil
}
