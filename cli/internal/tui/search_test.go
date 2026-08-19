package tui

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/jackjustus/credittracker/cli/internal/gen/client"
)

// fakeSearcher records the queries it was asked for and replays canned answers.
type fakeSearcher struct {
	queries  []string
	coasters []client.Coaster
	logged   []client.UUID
	err      error
}

func (f *fakeSearcher) SearchCoasterWithResponse(_ context.Context, params *client.SearchCoasterParams, _ ...client.RequestEditorFn) (*client.SearchCoasterResponse, error) {
	f.queries = append(f.queries, params.Q)
	if f.err != nil {
		return nil, f.err
	}
	coasters := f.coasters
	return &client.SearchCoasterResponse{
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
		JSON200:      &coasters,
	}, nil
}

func (f *fakeSearcher) LogRideWithResponse(_ context.Context, coasterID client.UUID, params *client.LogRideParams, _ ...client.RequestEditorFn) (*client.LogRideResponse, error) {
	f.logged = append(f.logged, coasterID)
	if f.err != nil {
		return nil, f.err
	}
	ride := client.Ride{
		Coaster:  client.Coaster{Id: coasterID, Name: "Kingda Ka"},
		RiddenAt: params.Timestamp,
	}
	return &client.LogRideResponse{
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
		JSON200:      &ride,
	}, nil
}

func newTestModel(t *testing.T, f *fakeSearcher) model {
	t.Helper()
	return newModel(t.Context(), f, "")
}

// update applies msg and returns the concrete model plus the command it emitted.
func update(t *testing.T, m model, msg tea.Msg) (model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(msg)
	got, ok := next.(model)
	if !ok {
		t.Fatalf("Update returned %T, want model", next)
	}
	return got, cmd
}

func typeRunes(t *testing.T, m model, s string) (model, tea.Cmd) {
	t.Helper()
	return update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)})
}

func TestTypingSearchesAfterDebounce(t *testing.T) {
	f := &fakeSearcher{coasters: []client.Coaster{
		{Id: uuid.New(), Name: "Kingda Ka"},
		{Id: uuid.New(), Name: "El Toro"},
	}}
	m := newTestModel(t, f)

	m, _ = typeRunes(t, m, "kingda")
	if len(f.queries) != 0 {
		t.Fatalf("searched before the debounce elapsed: %v", f.queries)
	}

	// Stand in for the timer firing.
	m, cmd := update(t, m, debouncedMsg{gen: m.gen})
	if cmd == nil {
		t.Fatal("debouncedMsg for the current generation did not start a search")
	}
	m, _ = update(t, m, cmd())

	if want := []string{"kingda"}; !equal(f.queries, want) {
		t.Errorf("queries = %v, want %v", f.queries, want)
	}
	if len(m.coasters) != 2 {
		t.Fatalf("coasters = %d, want 2", len(m.coasters))
	}
	if view := m.View(); !strings.Contains(view, "Kingda Ka") {
		t.Errorf("view does not list the results:\n%s", view)
	}
}

func TestStaleDebounceAndResultsAreDropped(t *testing.T) {
	f := &fakeSearcher{coasters: []client.Coaster{{Id: uuid.New(), Name: "Kingda Ka"}}}
	m := newTestModel(t, f)

	m, _ = typeRunes(t, m, "king")
	stale := m.gen
	m, _ = typeRunes(t, m, "da") // supersedes the first keystroke

	if _, cmd := update(t, m, debouncedMsg{gen: stale}); cmd != nil {
		t.Error("a superseded debounce still fired a search")
	}

	m, _ = update(t, m, resultsMsg{gen: stale, coasters: f.coasters})
	if len(m.coasters) != 0 {
		t.Errorf("applied results from a superseded query: %v", m.coasters)
	}
}

func TestEnterLogsSelectedCoaster(t *testing.T) {
	first, second := uuid.New(), uuid.New()
	f := &fakeSearcher{coasters: []client.Coaster{
		{Id: first, Name: "El Toro"},
		{Id: second, Name: "Kingda Ka"},
	}}
	m := newTestModel(t, f)

	m, _ = typeRunes(t, m, "kingda")
	m, _ = update(t, m, resultsMsg{gen: m.gen, coasters: f.coasters})
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyDown})

	m, cmd := update(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter did not log a ride")
	}
	m, _ = update(t, m, cmd())

	if len(f.logged) != 1 || f.logged[0] != second {
		t.Fatalf("logged = %v, want [%v]", f.logged, second)
	}
	if view := m.View(); !strings.Contains(view, "logged Kingda Ka") {
		t.Errorf("view does not confirm the ride:\n%s", view)
	}
}

func TestEnterWithNoResultsDoesNothing(t *testing.T) {
	f := &fakeSearcher{}
	m := newTestModel(t, f)

	if _, cmd := update(t, m, tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Error("enter acted on an empty result list")
	}
	if len(f.logged) != 0 {
		t.Errorf("logged %v with nothing selected", f.logged)
	}
}

func TestClearingQueryClearsResults(t *testing.T) {
	f := &fakeSearcher{coasters: []client.Coaster{{Id: uuid.New(), Name: "Kingda Ka"}}}
	m := newTestModel(t, f)

	m, _ = typeRunes(t, m, "kingda")
	m, _ = update(t, m, resultsMsg{gen: m.gen, coasters: f.coasters})

	for range "kingda" {
		m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	if len(m.coasters) != 0 {
		t.Errorf("coasters = %v, want none once the query is empty", m.coasters)
	}
	if view := m.View(); !strings.Contains(view, "type to search") {
		t.Errorf("view does not return to the empty state:\n%s", view)
	}
}

func TestSearchErrorIsShown(t *testing.T) {
	f := &fakeSearcher{err: errors.New("connection refused")}
	m := newTestModel(t, f)

	m, _ = typeRunes(t, m, "kingda")
	_, cmd := update(t, m, debouncedMsg{gen: m.gen})
	m, _ = update(t, m, cmd())

	if view := m.View(); !strings.Contains(view, "connection refused") {
		t.Errorf("view does not surface the search error:\n%s", view)
	}
}

func TestSeededQuerySearchesImmediately(t *testing.T) {
	f := &fakeSearcher{coasters: []client.Coaster{{Id: uuid.New(), Name: "Kingda Ka"}}}
	m := newModel(t.Context(), f, "kingda ka")

	if m.Init() == nil {
		t.Fatal("a seeded query did not search on start")
	}
	m, _ = update(t, m, search(t.Context(), f, m.gen, m.query)())

	if want := []string{"kingda ka"}; !equal(f.queries, want) {
		t.Errorf("queries = %v, want %v", f.queries, want)
	}
	if len(m.coasters) != 1 {
		t.Errorf("coasters = %d, want 1", len(m.coasters))
	}
}

func equal(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
