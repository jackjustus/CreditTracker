package handler

import (
	"context"

	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
)

func (s Server) GetCoaster(ctx context.Context, req api.SearchCoasterRequestObject) (api.SearchCoasterResponseObject, error) {
	return api.SearchCoaster200JSONResponse{}, nil
}
