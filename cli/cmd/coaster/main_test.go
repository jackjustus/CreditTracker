package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// resolve runs target resolution the way a real invocation does, through flag
// parsing, rather than calling resolveTarget with a hand-built command.
func resolve(t *testing.T, args ...string) (target, error) {
	t.Helper()

	// The flags read these, so a developer's own exported COASTER_TARGET would
	// otherwise decide the result. t.Setenv registers the restore; Unsetenv then
	// actually removes it, which t.Setenv alone cannot do.
	for _, key := range []string{"COASTER_SERVER", "COASTER_TARGET"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unsetting %s: %v", key, err)
		}
	}

	var (
		got    target
		gotErr error
	)
	cmd := &cli.Command{
		Name:  "coaster",
		Flags: rootFlags(),
		Action: func(_ context.Context, c *cli.Command) error {
			got, gotErr = resolveTarget(c)
			return nil
		},
	}
	if err := cmd.Run(context.Background(), append([]string{"coaster"}, args...)); err != nil {
		t.Fatalf("running %v: %v", args, err)
	}
	return got, gotErr
}

func TestReadConfirmation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "y", input: "y\n"},
		{name: "yes", input: "yes\n"},
		{name: "uppercase Y", input: "Y\n"},
		{name: "uppercase YES", input: "YES\n"},
		{name: "surrounding space", input: "  y  \n"},
		// A terminal answer that never got its newline, e.g. "y" then Ctrl-D.
		{name: "no trailing newline", input: "y"},
		{name: "n declines", input: "n\n", wantErr: true},
		{name: "empty line declines", input: "\n", wantErr: true},
		{name: "immediate EOF declines", input: "", wantErr: true},
		// Nothing but an explicit yes may approve, so near-misses must decline.
		{name: "yep declines", input: "yep\n", wantErr: true},
		{name: "garbage declines", input: "sure why not\n", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := readConfirmation(strings.NewReader(tc.input))
			if tc.wantErr && err == nil {
				t.Errorf("readConfirmation(%q) = nil, want a decline", tc.input)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("readConfirmation(%q) = %v, want approval", tc.input, err)
			}
		})
	}
}

func TestResolveTarget(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    target
		wantErr bool
	}{{
		name: "defaults to local",
		want: target{baseURL: localServer, label: "local", local: true},
	}, {
		name: "target prod",
		args: []string{"--target", "prod"},
		want: target{baseURL: prodServer, label: "prod", local: false},
	}, {
		name: "short target flag",
		args: []string{"-t", "prod"},
		want: target{baseURL: prodServer, label: "prod", local: false},
	}, {
		name: "explicit local",
		args: []string{"--target", "local"},
		want: target{baseURL: localServer, label: "local", local: true},
	}, {
		// --server has to win, otherwise a tunnel or a teammate's box would need
		// a code change to reach.
		name: "server overrides target",
		args: []string{"--target", "prod", "--server", "http://localhost:8080"},
		want: target{baseURL: "http://localhost:8080", label: "http://localhost:8080", local: true},
	}, {
		name: "loopback ip counts as local",
		args: []string{"--server", "http://127.0.0.1:9999"},
		want: target{baseURL: "http://127.0.0.1:9999", label: "http://127.0.0.1:9999", local: true},
	}, {
		// The guard must fire for an arbitrary host: unknown is not assumed safe.
		name: "arbitrary server is not local",
		args: []string{"--server", "http://129.213.157.228:8080"},
		want: target{baseURL: "http://129.213.157.228:8080", label: "http://129.213.157.228:8080", local: false},
	}, {
		name:    "unknown target",
		args:    []string{"--target", "staging"},
		wantErr: true,
	}, {
		name:    "server without scheme",
		args:    []string{"--server", "129.213.157.228:8080"},
		wantErr: true,
	}, {
		name:    "server without host",
		args:    []string{"--server", "http://"},
		wantErr: true,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolve(t, tc.args...)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveTarget = %+v, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveTarget: %v", err)
			}
			if got != tc.want {
				t.Errorf("resolveTarget = %+v, want %+v", got, tc.want)
			}
		})
	}
}
