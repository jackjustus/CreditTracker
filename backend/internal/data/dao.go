package data

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackjustus/credittracker/backend/internal/gen/db"
	"github.com/jackjustus/credittracker/backend/internal/models"
)

// SentinelUserID is the fixed user ID used for the single local user until
// authentication is implemented.
var SentinelUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type DAO struct {
	dbc db.Querier
}

func (dao *DAO) CreateRideEvent(ctx context.Context, userID uuid.UUID, coasterID uuid.UUID, createdAt time.Time) error {
	err := dao.dbc.CreateRideEvent(ctx, db.CreateRideEventParams{
		UserID:    userID,
		CoasterID: coasterID,
		CreatedAt: pgtype.Timestamptz{Time: createdAt, Valid: true},
	})
	if err != nil {
		return err
	}
	return nil
}

func (dao *DAO) HydrateCoaster(ctx context.Context, coasterID uuid.UUID) (*models.Coaster, error) {
	row, err := dao.dbc.HydrateCoaster(ctx, coasterID)
	if err != nil {
		return nil, err
	}
	return models.NewCoaster(row.Coaster), nil
}

func (dao *DAO) GetCredits(ctx context.Context, userID uuid.UUID) (*models.Credits, error) {
	events, err := dao.dbc.GetRideEvents(ctx, userID)
	if err != nil {
		return nil, err
	}
	return models.NewCredits(events)
}

func (dao *DAO) GetRides(ctx context.Context, userID uuid.UUID) (models.Rides, error) {
	events, err := dao.dbc.GetRideEvents(ctx, userID)
	if err != nil {
		return nil, err
	}
	return models.NewRides(events), nil
}

func NewDAO(ctx context.Context) (*DAO, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil, errors.New("no DATABASE_URL set in env")
	}
	pgpool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	dbc := db.New(pgpool)
	return &DAO{dbc}, nil
}
