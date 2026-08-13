package handler

import (
	"context"

	"github.com/jackjustus/credittracker/backend/internal/data"
	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
)

// ListRides lists rides on coasters from most recent to least.
// TODO: implement paging
func (s Server) ListRides(ctx context.Context, req api.ListRidesRequestObject) (api.ListRidesResponseObject, error) {
	rides, err := s.dao.GetRides(ctx, data.SentinelUserID)
	if err != nil {
		return nil, err
	}
	return api.ListRides200JSONResponse(rides.ToAPI()), nil
}
