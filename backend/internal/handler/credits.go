package handler

import (
	"context"

	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
)

func (s Server) ListCredits(ctx context.Context, req api.ListCreditsRequestObject) (api.ListCreditsResponseObject, error) {
	creds, err := s.dao.GetCredits(ctx)
	if err != nil {
		return nil, err
	}

	return api.ListCredits200JSONResponse(creds.ToAPI()), nil
}
