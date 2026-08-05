package handler

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackjustus/credittracker/backend/internal/gen/db"
	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
	"github.com/jackjustus/credittracker/backend/internal/views"
)

func (s Server) LogRide(ctx context.Context, req api.LogRideRequestObject) (api.LogRideResponseObject, error) {

	err := s.dbc.CreateRideEvent(ctx, db.CreateRideEventParams{
		CoasterID: req.CoasterID,
		CreatedAt: pgtype.Timestamptz{Time: req.Params.Timestamp, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	row, err := s.dbc.HydrateCoaster(ctx, req.CoasterID)
	if err != nil {
		return nil, err
	}

	coaster := views.NewCoaster(row)
	return api.LogRide200JSONResponse(coaster.ToRide(req.Params.Timestamp)), nil
}
