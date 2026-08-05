package handler

import (
	"context"

	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"

	"github.com/labstack/echo/v4"
)

func (s Server) ListRides(ctx context.Context, params api.ListRidesRequestObject) (api.ListRidesResponseObject, error) {
	return nil, echo.ErrServiceUnavailable
}
