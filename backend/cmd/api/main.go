package main

import (
	"context"
	"fmt"

	"github.com/jackjustus/credittracker/backend/internal/data"
	"github.com/jackjustus/credittracker/backend/internal/embed"
	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
	"github.com/jackjustus/credittracker/backend/internal/handler"
	"github.com/labstack/echo/v4"
)

func main() {
	ctx := context.Background()
	dao, err := data.NewDAO(ctx)
	if err != nil {
		fmt.Print(err)
		return
	}
	embedder, err := embed.NewClient()
	if err != nil {
		fmt.Print(err)
		return
	}
	r := handler.NewServer(dao, embedder)
	e := echo.New()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		c.Logger().Errorf("%s %s: %v", c.Request().Method, c.Path(), err)
		e.DefaultHTTPErrorHandler(err, c)
	}

	api.RegisterHandlers(e, api.NewStrictHandler(
		r,
		[]api.StrictMiddlewareFunc{},
	))
	err = e.Start(":8080")
	if err != nil {
		fmt.Print(err)
		return
	}
}
