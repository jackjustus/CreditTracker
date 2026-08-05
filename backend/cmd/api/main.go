package main

import (
	"context"
	"fmt"
	"github.com/jackjustus/credittracker/backend/internal/gen/db"
	api "github.com/jackjustus/credittracker/backend/internal/gen/openapi"
	"github.com/jackjustus/credittracker/backend/internal/handler"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func main() {
	dbc, err := dbsetup(context.Background())
	if err != nil {
		fmt.Print(err)
		return
	}
	r := handler.NewServer(dbc)
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

const defaultDatabaseURL = "postgres://postgres:testtest@localhost:5432/postgres?sslmode=disable"

func dbsetup(ctx context.Context) (*db.Queries, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = defaultDatabaseURL
	}
	pgpool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	return db.New(pgpool), nil
}
