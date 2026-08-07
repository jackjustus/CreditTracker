package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/jackjustus/credittracker/cli/internal/gen/client"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "coaster",
		Usage: "Inspect logged roller coaster rides",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "server",
				Usage:   "base URL of the CreditTracker API",
				Value:   "http://localhost:8080",
				Sources: cli.EnvVars("COASTER_SERVER"),
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "creds",
				Usage:  "list every coaster ridden at least once",
				Action: listCredits,
			},
			{
				Name:  "rides",
				Usage: "list the most recent rides",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "limit",
						Usage: "maximum rides to return",
						Value: 50,
					},
				},
				Action: listRides,
			},
			{
				Name:      "log",
				Usage:     "log a ride on a coaster",
				ArgsUsage: "<coaster-id>",
				Flags: []cli.Flag{
					&cli.TimestampFlag{
						Name:  "at",
						Usage: "when the ride happened, e.g. 2026-08-06T13:45:00Z (defaults to now)",
						Config: cli.TimestampConfig{
							Timezone: time.Local,
							Layouts:  []string{time.RFC3339, "2006-01-02 15:04", time.DateOnly},
						},
					},
				},
				Action: logRide,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "coaster:", err)
		os.Exit(1)
	}
}

func listCredits(ctx context.Context, cmd *cli.Command) error {
	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	resp, err := c.ListCreditsWithResponse(ctx)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("listing credits: %s", resp.Status())
	}

	return table([]string{"COASTER", "RIDES", "FIRST", "LAST"}, func(w *tabwriter.Writer) {
		for _, credit := range *resp.JSON200 {
			_, _ = fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
				credit.Name,
				credit.RideCount,
				credit.FirstRiddenAt.Format(time.DateOnly),
				credit.LastRiddenAt.Format(time.DateOnly))
		}
	})
}

func listRides(ctx context.Context, cmd *cli.Command) error {
	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	limit := int(cmd.Int("limit"))
	resp, err := c.ListRidesWithResponse(ctx, &client.ListRidesParams{Limit: &limit})
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("listing rides: %s", resp.Status())
	}

	return table([]string{"RIDDEN AT", "COASTER"}, func(w *tabwriter.Writer) {
		for _, ride := range resp.JSON200.Rides {
			_, _ = fmt.Fprintf(w, "%s\t%s\n",
				ride.RiddenAt.Format(time.RFC3339),
				ride.Coaster.Name)
		}
	})
}

func logRide(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("usage: coaster log <coaster-id>")
	}
	coasterID, err := uuid.Parse(cmd.Args().First())
	if err != nil {
		return fmt.Errorf("parsing coaster id: %w", err)
	}

	riddenAt := cmd.Timestamp("at")
	if riddenAt.IsZero() {
		riddenAt = time.Now()
	}

	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	resp, err := c.LogRideWithResponse(ctx, coasterID, &client.LogRideParams{Timestamp: riddenAt})
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("logging ride: %s", resp.Status())
	}

	ride := *resp.JSON200
	return table([]string{"RIDDEN AT", "COASTER"}, func(w *tabwriter.Writer) {
		_, _ = fmt.Fprintf(w, "%s\t%s\n",
			ride.RiddenAt.Format(time.RFC3339),
			ride.Coaster.Name)
	})
}

func newClient(cmd *cli.Command) (*client.ClientWithResponses, error) {
	return client.NewClientWithResponses(cmd.Root().String("server"))
}

func table(header []string, rows func(*tabwriter.Writer)) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	// tabwriter buffers writes, so any write error surfaces from Flush below.
	_, _ = fmt.Fprintln(w, strings.Join(header, "\t"))
	rows(w)
	return w.Flush()
}
