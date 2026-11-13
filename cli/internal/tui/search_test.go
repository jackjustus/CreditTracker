package tui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func (f *fakeSearcher) LogRideWithResponse(_ context.Context, coaster client.CoasterRef, params *client.LogRideParams, _ ...client.RequestEditorFn) (*client.LogRideResponse, error) {
	f.logged = append(f.logged, coaster.Id)
	if f.err != nil {
		return nil, f.err
	}
	ride := client.Ride{
		Coaster:  coaster,
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

func manyCoasters(n int) []client.Coaster {
	cs := make([]client.Coaster, n)
	for i := range cs {
		cs[i] = client.Coaster{Id: uuid.New(), Name: fmt.Sprintf("Coaster %02d", i)}
	}
	return cs
}

// Bubble Tea drops the top of a frame taller than the terminal, so a view that
// overflows hides the query line first.
func TestShortTerminalKeepsQueryVisible(t *testing.T) {
	f := &fakeSearcher{coasters: manyCoasters(10)}
	m := newModel(t.Context(), f, "kingda ka")
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 80, Height: 10})
	m, _ = update(t, m, search(t.Context(), f, m.gen, m.query)())

	view := m.View()
	if lines := strings.Count(view, "\n") + 1; lines > 10 {
		t.Errorf("view is %d lines, taller than the 10-line terminal:\n%s", lines, view)
	}
	if !strings.Contains(view, "> kingda ka") {
		t.Errorf("view does not show the query:\n%s", view)
	}
}

func TestCursorScrollsThroughResults(t *testing.T) {
	f := &fakeSearcher{coasters: manyCoasters(10)}
	m := newModel(t.Context(), f, "kingda ka")
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 80, Height: 10}) // 4 result rows
	m, _ = update(t, m, search(t.Context(), f, m.gen, m.query)())

	for range 9 {
		m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyDown})
	}
	view := m.View()
	if !strings.Contains(view, "Coaster 09") || strings.Contains(view, "Coaster 05") {
		t.Errorf("bottom of the list is not scrolled into view:\n%s", view)
	}
	if !strings.Contains(view, "10/10") {
		t.Errorf("view does not show the position in the list:\n%s", view)
	}

	for range 9 {
		m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyUp})
	}
	if view := m.View(); !strings.Contains(view, "Coaster 00") {
		t.Errorf("top of the list is not scrolled back into view:\n%s", view)
	}
}

func TestLongNamesDoNotWrap(t *testing.T) {
	f := &fakeSearcher{coasters: []client.Coaster{{Id: uuid.New(), Name: strings.Repeat("x", 100)}}}
	m := newModel(t.Context(), f, "kingda ka")
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 40, Height: 20})
	m, _ = update(t, m, search(t.Context(), f, m.gen, m.query)())

	for _, line := range strings.Split(m.View(), "\n") {
		if w := lipgloss.Width(line); w > 40 {
			t.Errorf("line is %d columns wide, wider than the 40-column terminal: %q", w, line)
		}
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
