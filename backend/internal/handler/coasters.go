package handler

import (
	"context"

	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
)

const searchLimit = 10

// SearchCoaster embeds the query returns the nearest coasters.
func (s Server) SearchCoaster(ctx context.Context, req api.SearchCoasterRequestObject) (api.SearchCoasterResponseObject, error) {
	query, err := s.embedder.EmbedOne(ctx, req.Params.Q)
	if err != nil {
		return nil, err
	}

	coasters, err := s.dao.SearchCoasters(ctx, query, searchLimit)
	if err != nil {
		return nil, err
	}

	return api.SearchCoaster200JSONResponse(coasters.ToAPI()), nil
}
