// Package tui holds the interactive terminal front-ends for the coaster CLI.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackjustus/credittracker/cli/internal/gen/client"
)

// searchDebounce is how long typing must pause before a query is sent. Every
// search embeds the query server-side, so a request per keystroke is wasteful.
const searchDebounce = 250 * time.Millisecond

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	promptStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	okStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
)

// Searcher is the slice of the generated client the search TUI needs.
type Searcher interface {
	SearchCoasterWithResponse(ctx context.Context, params *client.SearchCoasterParams, reqEditors ...client.RequestEditorFn) (*client.SearchCoasterResponse, error)
	LogRideWithResponse(ctx context.Context, coasterID client.UUID, params *client.LogRideParams, reqEditors ...client.RequestEditorFn) (*client.LogRideResponse, error)
}

// debouncedMsg fires once typing has been idle for searchDebounce.
type debouncedMsg struct{ gen int }

// resultsMsg carries the outcome of one search request.
type resultsMsg struct {
	gen      int
	coasters []client.Coaster
	err      error
}

// loggedMsg carries the outcome of one ride log.
type loggedMsg struct {
	ride client.Ride
	err  error
}

type model struct {
	ctx    context.Context
	client Searcher

	input    textinput.Model
	coasters []client.Coaster
	cursor   int

	// gen increments on every edit so replies to superseded queries are dropped.
	gen       int
	inflight  bool
	query     string
	searchErr error
	logged    string
	logErr    error
}

// RunSearch starts the interactive coaster search, seeded with query.
func RunSearch(ctx context.Context, c Searcher, query string) error {
	_, err := tea.NewProgram(newModel(ctx, c, query), tea.WithContext(ctx)).Run()
	return err
}

func newModel(ctx context.Context, c Searcher, query string) model {
	query = strings.TrimSpace(query)

	in := textinput.New()
	in.Placeholder = "kingda ka"
	in.Prompt = "> "
	in.PromptStyle = promptStyle
	in.SetValue(query)
	in.CursorEnd()
	in.Focus()

	return model{ctx: ctx, client: c, input: in, query: query}
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{textinput.Blink}
	if m.query != "" {
		// Seeded from the command line: search immediately, no debounce.
		cmds = append(cmds, search(m.ctx, m.client, m.gen, m.query))
	}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyUp, tea.KeyCtrlP:
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case tea.KeyDown, tea.KeyCtrlN:
			if m.cursor < len(m.coasters)-1 {
				m.cursor++
			}
			return m, nil
		case tea.KeyEnter:
			if m.cursor >= len(m.coasters) {
				return m, nil
			}
			m.logged, m.logErr = "", nil
			return m, logRide(m.ctx, m.client, m.coasters[m.cursor])
		}

	case debouncedMsg:
		// A newer keystroke already bumped gen; this timer is stale.
		if msg.gen != m.gen {
			return m, nil
		}
		if m.query == "" {
			return m, nil
		}
		m.inflight = true
		return m, search(m.ctx, m.client, m.gen, m.query)

	case resultsMsg:
		if msg.gen != m.gen {
			return m, nil
		}
		m.inflight = false
		m.searchErr = msg.err
		if msg.err == nil {
			m.coasters = msg.coasters
			if m.cursor >= len(m.coasters) {
				m.cursor = max(len(m.coasters)-1, 0)
			}
		}
		return m, nil

	case loggedMsg:
		if msg.err != nil {
			m.logErr = msg.err
			return m, nil
		}
		m.logged = fmt.Sprintf("logged %s at %s",
			msg.ride.Coaster.Name, msg.ride.RiddenAt.Local().Format("Jan 2 15:04"))
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	if q := strings.TrimSpace(m.input.Value()); q != m.query {
		m.query = q
		m.gen++
		m.searchErr = nil
		if q == "" {
			m.coasters, m.cursor, m.inflight = nil, 0, false
			return m, cmd
		}
		return m, tea.Batch(cmd, debounce(m.gen))
	}
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("coaster search"))
	b.WriteString("\n\n")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")

	switch {
	case m.searchErr != nil:
		b.WriteString(errorStyle.Render("  " + m.searchErr.Error()))
	case m.query == "":
		b.WriteString(dimStyle.Render("  type to search"))
	case len(m.coasters) == 0 && m.inflight:
		b.WriteString(dimStyle.Render("  searching…"))
	case len(m.coasters) == 0:
		b.WriteString(dimStyle.Render("  no matches"))
	default:
		for i, c := range m.coasters {
			if i == m.cursor {
				b.WriteString(selectedStyle.Render("❯ " + c.Name))
			} else {
				b.WriteString("  " + c.Name)
			}
			b.WriteByte('\n')
		}
	}

	b.WriteString("\n\n")
	switch {
	case m.logErr != nil:
		b.WriteString(errorStyle.Render("  " + m.logErr.Error()))
	case m.logged != "":
		b.WriteString(okStyle.Render("  ✓ " + m.logged))
	default:
		b.WriteString(dimStyle.Render("  ↑/↓ move · enter log a ride · esc quit"))
	}
	b.WriteByte('\n')

	return b.String()
}

// debounce waits out searchDebounce before letting generation gen search.
func debounce(gen int) tea.Cmd {
	return tea.Tick(searchDebounce, func(time.Time) tea.Msg {
		return debouncedMsg{gen: gen}
	})
}

func search(ctx context.Context, c Searcher, gen int, query string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.SearchCoasterWithResponse(ctx, &client.SearchCoasterParams{Q: query})
		switch {
		case err != nil:
			return resultsMsg{gen: gen, err: err}
		case resp.JSON200 == nil:
			return resultsMsg{gen: gen, err: fmt.Errorf("searching coasters: %s", resp.Status())}
		}
		return resultsMsg{gen: gen, coasters: *resp.JSON200}
	}
}

func logRide(ctx context.Context, c Searcher, coaster client.Coaster) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.LogRideWithResponse(ctx, coaster.Id, &client.LogRideParams{Timestamp: time.Now()})
		switch {
		case err != nil:
			return loggedMsg{err: err}
		case resp.JSON200 == nil:
			return loggedMsg{err: fmt.Errorf("logging ride: %s", resp.Status())}
		}
		return loggedMsg{ride: *resp.JSON200}
	}
}
