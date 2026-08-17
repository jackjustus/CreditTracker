package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/jackjustus/credittracker/cli/internal/gen/client"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

const (
	localServer = "http://localhost:8080"
	// The A1's public IP. Stable for the life of the instance, but not reserved:
	// recreating the instance moves it and this constant has to follow.
	prodServer = "http://129.213.157.228:8080"
)

// rootFlags is shared with the tests, which exercise target resolution without
// standing up the whole command tree.
func rootFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "server",
			Usage:   "base URL of the CreditTracker API (overrides --target)",
			Sources: cli.EnvVars("COASTER_SERVER"),
		},
		&cli.StringFlag{
			Name:    "target",
			Aliases: []string{"t"},
			Usage:   `which deployment to talk to: "local" or "prod"`,
			Value:   "local",
			Sources: cli.EnvVars("COASTER_TARGET"),
		},
	}
}

func main() {
	cmd := &cli.Command{
		Name:  "coaster",
		Usage: "Inspect logged roller coaster rides",
		Flags: rootFlags(),
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
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "skip the confirmation prompt on a non-local target",
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
	c, _, err := newClient(cmd)
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
	c, _, err := newClient(cmd)
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
		for _, ride := range *resp.JSON200 {
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

	c, t, err := newClient(cmd)
	if err != nil {
		return err
	}
	if err := confirmWrite(cmd, t); err != nil {
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

// target is the API endpoint one invocation resolved to.
type target struct {
	baseURL string
	label   string
	local   bool
}

// resolveTarget picks the endpoint from --server, then --target, then the
// default. Local is the default deliberately: forgetting the flag has to be the
// harmless outcome.
func resolveTarget(cmd *cli.Command) (target, error) {
	// --server wins outright, so a one-off host (an ssh tunnel, someone else's
	// box) needs no code change. Its hostname decides whether the write guard
	// applies, since an arbitrary URL is not assumed to be safe.
	if server := cmd.Root().String("server"); server != "" {
		u, err := parseServer(server)
		if err != nil {
			return target{}, err
		}
		return target{baseURL: server, label: server, local: isLoopback(u.Hostname())}, nil
	}

	switch name := cmd.Root().String("target"); name {
	case "local":
		return target{baseURL: localServer, label: "local", local: true}, nil
	case "prod":
		return target{baseURL: prodServer, label: "prod", local: false}, nil
	default:
		return target{}, fmt.Errorf("unknown target %q: want \"local\" or \"prod\"", name)
	}
}

// parseServer rejects what url.Parse accepts but the HTTP client cannot use, so
// a typo fails here rather than as a confusing transport error later.
func parseServer(server string) (*url.URL, error) {
	u, err := url.Parse(server)
	if err != nil {
		return nil, fmt.Errorf("parsing --server %q: %w", server, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("--server %q: want an http:// or https:// URL", server)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("--server %q: missing host", server)
	}
	return u, nil
}

func isLoopback(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	return false
}

func newClient(cmd *cli.Command) (*client.ClientWithResponses, target, error) {
	t, err := resolveTarget(cmd)
	if err != nil {
		return nil, target{}, err
	}
	// Local is both the default and the safe case, so it stays quiet. Anything
	// else announces itself: prod and local hold the same catalog and their
	// output is indistinguishable.
	if !t.local {
		_, _ = fmt.Fprintf(os.Stderr, "coaster: targeting %s (%s)\n", t.label, t.baseURL)
	}
	c, err := client.NewClientWithResponses(t.baseURL)
	if err != nil {
		return nil, target{}, err
	}
	return c, t, nil
}

// confirmWrite gates mutations against anything but a local server. Reads are
// harmless wherever they run, so only `log` calls this.
func confirmWrite(cmd *cli.Command, t target) error {
	if t.local || cmd.Bool("yes") {
		return nil
	}

	// Fail closed with no terminal to prompt at: a script that forgot its target
	// must not write to prod just because nobody was there to say no. This is an
	// ioctl, not a check on the file mode -- /dev/null is a character device too.
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("refusing to log a ride to %s without a terminal to confirm at: pass --yes", t.label)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Log a ride to %s (%s)? [y/N] ", t.label, t.baseURL)
	return readConfirmation(os.Stdin)
}

// readConfirmation treats anything but an explicit yes as a decline, so a stray
// keystroke cannot be an approval.
func readConfirmation(r io.Reader) error {
	answer, err := bufio.NewReader(r).ReadString('\n')
	// EOF is not a failure to read: Ctrl-D on an empty line leaves answer empty
	// and declines below, while "y" then Ctrl-D is a real answer with no newline.
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("reading confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return nil
	}
	return errors.New("aborted")
}

func table(header []string, rows func(*tabwriter.Writer)) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	// tabwriter buffers writes, so any write error surfaces from Flush below.
	_, _ = fmt.Fprintln(w, strings.Join(header, "\t"))
	rows(w)
	return w.Flush()
}
