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
	LogRideWithResponse(ctx context.Context, coaster client.CoasterRef, params *client.LogRideParams, reqEditors ...client.RequestEditorFn) (*client.LogRideResponse, error)
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
	// name comes from the selected result: Ride only carries a coaster ref.
	name string
	ride client.Ride
	err  error
}

type model struct {
	ctx    context.Context
	client Searcher

	input    textinput.Model
	coasters []client.Coaster
	cursor   int
	// offset is the first result on screen when they do not all fit.
	offset int

	// width and height are the terminal size; zero until the first
	// WindowSizeMsg, which means "unbounded".
	width, height int

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
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		// Scroll a long query sideways instead of wrapping it: prompt + cursor.
		m.input.Width = max(msg.Width-lipgloss.Width(m.input.Prompt)-1, 1)
		m.scrollToCursor()
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyUp, tea.KeyCtrlP:
			if m.cursor > 0 {
				m.cursor--
			}
			m.scrollToCursor()
			return m, nil
		case tea.KeyDown, tea.KeyCtrlN:
			if m.cursor < len(m.coasters)-1 {
				m.cursor++
			}
			m.scrollToCursor()
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
			m.scrollToCursor()
		}
		return m, nil

	case loggedMsg:
		if msg.err != nil {
			m.logErr = msg.err
			return m, nil
		}
		m.logged = fmt.Sprintf("logged %s at %s",
			msg.name, msg.ride.RiddenAt.Local().Format("Jan 2 15:04"))
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	if q := strings.TrimSpace(m.input.Value()); q != m.query {
		m.query = q
		m.gen++
		m.searchErr = nil
		if q == "" {
			m.coasters, m.cursor, m.offset, m.inflight = nil, 0, 0, false
			return m, cmd
		}
		return m, tea.Batch(cmd, debounce(m.gen))
	}
	return m, cmd
}

// chromeLines is every line View draws besides the results: title, blank,
// input, blank, blank, footer.
const chromeLines = 6

// resultRows is how many results fit on screen. Bubble Tea drops the top of a
// frame taller than the terminal, which would hide the query, so the list
// gives way instead.
func (m model) resultRows() int {
	if m.height == 0 {
		return len(m.coasters)
	}
	return max(m.height-chromeLines, 1)
}

// scrollToCursor moves the visible window just enough to keep the cursor in it.
func (m *model) scrollToCursor() {
	rows := m.resultRows()
	switch {
	case m.cursor < m.offset:
		m.offset = m.cursor
	case m.cursor >= m.offset+rows:
		m.offset = m.cursor - rows + 1
	}
	m.offset = max(min(m.offset, len(m.coasters)-rows), 0)
}

func (m model) View() string {
	lines := []string{
		titleStyle.Render("coaster search"),
		"",
		m.input.View(),
		"",
	}

	switch {
	case m.searchErr != nil:
		lines = append(lines, errorStyle.Render("  "+m.searchErr.Error()))
	case m.query == "":
		lines = append(lines, dimStyle.Render("  type to search"))
	case len(m.coasters) == 0 && m.inflight:
		lines = append(lines, dimStyle.Render("  searching…"))
	case len(m.coasters) == 0:
		lines = append(lines, dimStyle.Render("  no matches"))
	default:
		// A wrapped name would take two rows and push the query off the top.
		row := lipgloss.NewStyle()
		if m.width > 0 {
			row = row.MaxWidth(m.width)
		}
		end := min(m.offset+m.resultRows(), len(m.coasters))
		for i := m.offset; i < end; i++ {
			name := m.coasters[i].Name
			if i == m.cursor {
				lines = append(lines, row.Render(selectedStyle.Render("❯ "+name)))
			} else {
				lines = append(lines, row.Render("  "+name))
			}
		}
	}

	lines = append(lines, "")
	switch {
	case m.logErr != nil:
		lines = append(lines, errorStyle.Render("  "+m.logErr.Error()))
	case m.logged != "":
		lines = append(lines, okStyle.Render("  ✓ "+m.logged))
	default:
		help := "  ↑/↓ move · enter log a ride · esc quit"
		if n := len(m.coasters); m.resultRows() < n {
			help = fmt.Sprintf("  %d/%d · ↑/↓ move · enter log a ride · esc quit", m.cursor+1, n)
		}
		lines = append(lines, dimStyle.Render(help))
	}

	return strings.Join(lines, "\n")
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
		resp, err := c.LogRideWithResponse(ctx, client.CoasterRef{Id: coaster.Id}, &client.LogRideParams{Timestamp: time.Now()})
		switch {
		case err != nil:
			return loggedMsg{err: err}
		case resp.JSON200 == nil:
			return loggedMsg{err: fmt.Errorf("logging ride: %s", resp.Status())}
		}
		return loggedMsg{name: coaster.Name, ride: *resp.JSON200}
	}
}
