package handler

import (
	"context"

	"github.com/jackjustus/credittracker/backend/internal/data"
	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
)

func (s Server) LogRide(ctx context.Context, req api.LogRideRequestObject) (api.LogRideResponseObject, error) {

	err := s.dao.CreateRideEvent(ctx, data.SentinelUserID, req.CoasterID, req.Params.Timestamp)
	if err != nil {
		return nil, err
	}

	coaster, err := s.dao.HydrateCoaster(ctx, req.CoasterID)
	if err != nil {
		return nil, err
	}

	return api.LogRide200JSONResponse(coaster.ToRide(req.Params.Timestamp)), nil
}
