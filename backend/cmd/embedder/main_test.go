package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackjustus/credittracker/backend/internal/gen/db"
	"github.com/jackjustus/credittracker/backend/internal/models"
)

func TestCoasterDoc(t *testing.T) {
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	parkID := uuid.New()

	c := models.NewCoaster(db.Coaster{
		ID:             uuid.New(),
		ParkID:         parkID,
		Name:           "Kingda Ka",
		ManufacturedAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	p := models.NewPark(db.Park{
		ID:        parkID,
		Name:      "Six Flags Great Adventure",
		City:      "Jackson",
		Country:   "United States",
		CreatedAt: now,
		UpdatedAt: now,
	})
	cwp, err := models.NewCoasterWithPark(c, p)
	if err != nil {
		t.Fatalf("NewCoasterWithPark: %v", err)
	}

	doc, err := CoasterDoc(*cwp)
	if err != nil {
		t.Fatalf("CoasterDoc: %v", err)
	}

	// Pinned exactly: this string is what gets embedded, so a wording change
	// invalidates every vector already stored under the old phrasing.
	want := "Kingda Ka is a coaster at Six Flags Great Adventure"
	if doc != want {
		t.Errorf("CoasterDoc = %q, want %q", doc, want)
	}
}
